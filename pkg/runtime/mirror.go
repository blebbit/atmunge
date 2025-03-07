package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	plcdb "github.com/blebbit/plc-mirror/pkg/db"
	"github.com/blebbit/plc-mirror/pkg/plc"
)

func (r *Runtime) StartMirror() {
	log := zerolog.Ctx(r.ctx).With().Str("module", "mirror").Logger()
	for {
		select {
		case <-r.ctx.Done():
			log.Info().Msgf("PLC mirror stopped")
			return
		default:
			if err := r.backfillMirror(); err != nil {
				if r.ctx.Err() == nil {
					log.Error().Err(err).Msgf("Failed to get new log entries from PLC: %s", err)
				}
			} else {
				now := time.Now()
				r.mu.Lock()
				r.lastCompletionTimestamp = now
				r.mu.Unlock()
			}
			time.Sleep(10 * time.Second)
		}
	}
}

func (r *Runtime) backfillMirror() error {
	log := zerolog.Ctx(r.ctx)

	cursor := ""
	err := r.db.Model(&plcdb.PLCLogEntry{}).Select("plc_timestamp").Order("plc_timestamp desc").Limit(1).Take(&cursor).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get the cursor: %w", err)
	}

	cursorTimestamp, err := time.Parse(time.RFC3339, cursor)
	if err != nil {
		log.Error().Err(err).Msgf("parsing timestamp %q: %s", cursor, err)
	} else {
		r.updateRateLimit(cursorTimestamp)
	}

	u := *r.upstream

	// loop to get 1000 records at a time until we are caught up
	for {
		params := u.Query()
		params.Set("count", "1000")
		if cursor != "" {
			params.Set("after", cursor)
		}
		u.RawQuery = params.Encode()

		req, err := http.NewRequestWithContext(r.ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return fmt.Errorf("constructing request: %w", err)
		}

		_ = r.limiter.Wait(r.ctx)
		log.Info().Msgf("Listing PLC log entries with cursor %q...", cursor)
		log.Debug().Msgf("Request URL: %s", u.String())
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("sending request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		newEntries := []plcdb.PLCLogEntry{}
		mapInfos := map[string]plcdb.AccountInfo{}
		decoder := json.NewDecoder(resp.Body)
		oldCursor := cursor

		var lastTimestamp time.Time

		// decode each jsonl line
		for {
			var entry plc.OperationLogEntry
			err := decoder.Decode(&entry)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return fmt.Errorf("parsing log entry: %w", err)
			}

			// turn entry into DB types
			row := plcdb.PLCLogEntryFromOp(entry)
			info := plcdb.AccountInfoFromOp(entry)

			// update lastestTimestamp / cursor
			t, err := time.Parse(time.RFC3339, row.PLCTimestamp)
			if err == nil {
				lastEventTimestamp.Set(float64(t.Unix()))
				lastTimestamp = t
			} else {
				log.Warn().Msgf("Failed to parse %q: %s", row.PLCTimestamp, err)
			}
			cursor = entry.CreatedAt

			// skip bogus records
			if info.PDS == "https://uwu" {
				continue
			}

			// TODO: validate _atproto.<handle> points at same DID
			// ... or be lazy about it (probably better choice) ...

			// add to tmp collections
			mapInfos[info.DID] = info
			newEntries = append(newEntries, row)
		}

		// check if we are caught up, end inf loop if so
		if len(newEntries) == 0 || cursor == oldCursor {
			break
		}

		// write PLC Log rows
		err = r.db.Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "did"}, {Name: "cid"}},
				DoNothing: true,
			},
		).Create(newEntries).Error
		if err != nil {
			return fmt.Errorf("inserting log entry into database: %w", err)
		}

		// write Acct Info rows
		newInfos := make([]plcdb.AccountInfo, 0, len(mapInfos))
		for _, v := range mapInfos {
			newInfos = append(newInfos, v)
		}
		err = r.db.Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "did"}},
				DoUpdates: clause.AssignmentColumns([]string{"plc_timestamp", "pds", "handle"}),
			},
		).Create(newInfos).Error
		if err != nil {
			return fmt.Errorf("inserting acct info into database: %w", err)
		}

		// update tiemstamp & rate-limiter
		if !lastTimestamp.IsZero() {
			r.mu.Lock()
			r.lastRecordTimestamp = lastTimestamp
			r.mu.Unlock()

			r.updateRateLimit(lastTimestamp)
		}

		log.Info().Msgf("Got %d | %d log entries. New cursor: %q", len(newEntries), len(newInfos), cursor)
	}
	return nil
}
