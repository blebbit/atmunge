package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	atdb "github.com/blebbit/at-mirror/pkg/db"
	"github.com/blebbit/at-mirror/pkg/plc"
)

func (r *Runtime) StartPLCBackfill() {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "plc").Logger()
	for {
		select {
		case <-r.Ctx.Done():
			log.Info().Msgf("PLC backfill stopped")
			return
		default:
			if err := r.BackfillMirror(); err != nil {
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

func (r *Runtime) StartPLCMirror() {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "plc").Logger()
	for {
		select {
		case <-r.Ctx.Done():
			log.Info().Msgf("PLC mirror stopped")
			return
		default:
			if err := r.BackfillMirror(); err != nil {
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

func (r *Runtime) BackfillMirror() error {
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
			// if ok := validateOperation(entry, op); !ok {
			// 	continue
			// }

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

func (r *Runtime) BackfillPlcLogs() error {
	log := zerolog.Ctx(r.Ctx)

	log.Info().Msgf("Starting PLC log backfill...")
	jsonBytes, jerr := json.MarshalIndent(r.Cfg, "", "  ")
	if jerr != nil {
		fmt.Println("Error marshalling config to JSON:", jerr)
		return jerr
	}
	log.Info().Msgf(string(jsonBytes))
	// fmt.Println(string(jsonBytes))

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

	// bookeeping
	var good, bad, errs int

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
		log.Debug().Msgf("Listing PLC log entries with cursor %q...", cursor)
		log.Debug().Msgf("Request URL: %s", u.String())
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "context canceled") {
				log.Info().Msgf("PLC backfill stopped by context cancellation")
				return nil
			}
			log.Error().Err(err).Msgf("sending request: %s", err)
			errs++
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			log.Error().Err(err).Msgf("unexpected status code: %d", resp.StatusCode)
			errs++
			continue
		}

		newEntries := []atdb.PLCLogEntry{}
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
				if strings.Contains(err.Error(), "connection reset by peer") {
					log.Error().Err(err).Msgf("parsing log entry: %s", err)
					errs++
					time.Sleep(5 * time.Second) // wait a bit before retrying
					break
				}
				log.Error().Err(err).Msgf("parsing log entry: %s", err)
				bad++
				continue
			}

			// update cursor
			cursor = entry.CreatedAt

			// turn the entry into a PLC operation
			var op plc.Op
			switch v := entry.Operation.Value.(type) {
			case plc.Op:
				op = v
			case plc.LegacyCreateOp:
				op = v.AsUnsignedOp()
			}

			// turn entry into DB types
			row := atdb.PLCLogEntryFromOp(entry)

			// update lastestTimestamp / cursor
			t, err := time.Parse(time.RFC3339, row.PLCTimestamp)
			if err == nil {
				lastEventTimestamp.Set(float64(t.Unix()))
				lastTimestamp = t
			} else {
				log.Warn().Msgf("Failed to parse %q: %s", row.PLCTimestamp, err)
				errs++
			}

			// filter operations by various means
			if r.Cfg.PlcFilter {

				if !validateOperation(entry, op) {
					bad++
					continue
				}

				info := atdb.AccountInfoFromOp(entry)
				// skip bogus records
				if info.PDS == "https://uwu" {
					log.Warn().Msgf("Skipping entry with bogus PDS: %s", info.PDS)
					bad++
					continue
				}

				doc, err := plc.MakeDoc(entry, op)
				if err != nil {
					log.Debug().Err(err).Msgf("Failed to create DID document for entry %s: %s", entry.CID, err)
					bad++
					continue
				}
				docJSON, err := json.Marshal(doc)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to Marshal DID document for entry %s: %s", entry.CID, err)
					errs++
					continue
				}
				log.Debug().Msgf("DID Document for %s: %s", entry.DID, docJSON)
			}
			// TODO: validate _atproto.<handle> points at same DID
			// ... or be lazy about it (probably better choice) ...

			// add to tmp collections
			good++
			newEntries = append(newEntries, row)
		}

		// check if we are caught up, end inf loop if so
		if cursor == oldCursor {
			log.Warn().Msgf("Caught up with PLC logs, no new entries found. %s", cursor)
			break
		}

		// write PLC Log rows
		if len(newEntries) > 0 {
			err = r.DB.Clauses(
			// clause.OnConflict{
			// 	Columns: []clause.Column{{Name: "did"}, {Name: "cid"}},
			// 	// HMMM, what should we do on conflicts?
			// 	DoNothing: true,
			// 	// UpdateAll: true,
			// 	// (ideally), write to conflict table
			// },
			).Create(newEntries).Error
			if err != nil {
				return fmt.Errorf("inserting log entry into database: %w", err)
			}
		}

		// update timestamp & rate-limiter
		if !lastTimestamp.IsZero() {
			r.plcMutex.Lock()
			r.lastRecordTimestamp = lastTimestamp
			r.plcMutex.Unlock()

			r.updateRateLimit(lastTimestamp)
		}

		log.Info().Msgf("%d | %d | %d | %d | %d entries. New cursor: %q", good, bad, errs, good+bad+errs, len(newEntries), cursor)
	}

	return nil
}
