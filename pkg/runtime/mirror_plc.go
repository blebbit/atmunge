package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	atdb "github.com/blebbit/at-mirror/pkg/db"
	"github.com/blebbit/at-mirror/pkg/plc"
)

func (r *Runtime) StartPLCMirror() {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "plc").Logger()
	for {
		select {
		case <-r.Ctx.Done():
			log.Info().Msgf("PLC mirror stopped")
			return
		default:
			if err := r.backfillMirror(); err != nil {
				if r.Ctx.Err() == nil {
					log.Error().Err(err).Msgf("Failed to get new log entries from PLC: %s", err)
				}
			} else {
				now := time.Now()
				r.plcMutex.Lock()
				r.lastCompletionTimestamp = now
				r.plcMutex.Unlock()
			}
			time.Sleep(10 * time.Second)
		}
	}
}

func (r *Runtime) backfillMirror() error {
	log := zerolog.Ctx(r.Ctx)

	cursor := ""
	err := r.DB.Model(&atdb.PLCLogEntry{}).Select("plc_timestamp").Order("plc_timestamp desc").Limit(1).Take(&cursor).Error
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

		req, err := http.NewRequestWithContext(r.Ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return fmt.Errorf("constructing request: %w", err)
		}

		_ = r.limiter.Wait(r.Ctx)
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

		newEntries := []atdb.PLCLogEntry{}
		mapInfos := map[string]atdb.AccountInfo{}
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

			// turn the entry into a PLC operation
			var op plc.Op
			switch v := entry.Operation.Value.(type) {
			case plc.Op:
				op = v
			case plc.LegacyCreateOp:
				op = v.AsUnsignedOp()
			}

			// skip entries that are not valid operations
			if ok := validateOperation(entry, op); !ok {
				continue
			}

			doc, err := plc.MakeDoc(entry, op)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to create DID document for entry %s: %s", entry.CID, err)
				continue
			}
			docJSON, err := json.Marshal(doc)
			log.Debug().Msgf("DID Document for %s: %s", entry.DID, docJSON)

			// turn entry into DB types
			row := atdb.PLCLogEntryFromOp(entry)
			info := atdb.AccountInfoFromOp(entry)

			// update lastestTimestamp / cursor
			t, err := time.Parse(time.RFC3339, row.PLCTimestamp)
			if err == nil {
				lastEventTimestamp.Set(float64(t.Unix()))
				lastTimestamp = t
			} else {
				log.Warn().Msgf("Failed to parse %q: %s", row.PLCTimestamp, err)
			}
			cursor = entry.CreatedAt

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
		err = r.DB.Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "did"}, {Name: "cid"}},
				DoNothing: true,
			},
		).Create(newEntries).Error
		if err != nil {
			return fmt.Errorf("inserting log entry into database: %w", err)
		}

		// write Acct Info rows
		newInfos := make([]atdb.AccountInfo, 0, len(mapInfos))
		for _, v := range mapInfos {
			newInfos = append(newInfos, v)
		}
		err = r.DB.Clauses(
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
			r.plcMutex.Lock()
			r.lastRecordTimestamp = lastTimestamp
			r.plcMutex.Unlock()

			r.updateRateLimit(lastTimestamp)
		}

		log.Info().Msgf("Got %d | %d log entries. New cursor: %q", len(newEntries), len(newInfos), cursor)
	}

	return nil
}

func validateOperation(entry plc.OperationLogEntry, op plc.Op) bool {

	// check did
	if entry.DID == "" {
		return false
	}

	// check handle
	if len(op.AlsoKnownAs) > 0 {
		handle := strings.TrimPrefix(op.AlsoKnownAs[0], "at://")
		// check if handle is a valid handle
		if _, err := url.Parse(handle); err != nil {
			return false
		}
	}

	// check pds
	if svc, ok := op.Services["atproto_pds"]; ok {
		pds := svc.Endpoint
		if _, err := url.Parse(pds); err != nil {
			return false
		}

		// check some well-known bad values
		switch pds {

		}
	}

	return true
}
