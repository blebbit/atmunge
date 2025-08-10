package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	atdb "github.com/blebbit/at-mirror/pkg/db"
	"github.com/blebbit/at-mirror/pkg/plc"
)

// func (r *Runtime) StartPLCMirror() {
// 	log := zerolog.Ctx(r.Ctx).With().Str("module", "plc").Logger()
// 	for {
// 		select {
// 		case <-r.Ctx.Done():
// 			log.Info().Msgf("PLC mirror stopped")
// 			return
// 		default:
// 			if err := r.BackfillMirror(); err != nil {
// 				if r.Ctx.Err() == nil {
// 					log.Error().Err(err).Msgf("Failed to get new log entries from PLC: %s", err)
// 				}
// 			} else {
// 				now := time.Now()
// 				r.plcMutex.Lock()
// 				r.lastCompletionTimestamp = now
// 				r.plcMutex.Unlock()
// 			}
// 			time.Sleep(10 * time.Second)
// 		}
// 	}
// }

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
	err := r.DB.WithContext(r.Ctx).Model(&atdb.PLCLogEntry{}).
		Select("plc_timestamp").Order("plc_timestamp desc").Limit(1).Take(&cursor).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get the cursor: %w", err)
	}

	u, err := url.Parse(r.Cfg.PlcUpstream)
	if err != nil {
		return err
	}

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

		// log.Info().Msgf("Request URL: %s", u.String())

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
					break
				}
				if strings.Contains(err.Error(), "invalid character '<' looking") {
					log.Error().Err(err).Msgf("parsing log entry: %s", err)
					log.Error().Msgf("Try to manually fetch the logs from %q", u.String())
					errs++
					break
				}
				if strings.Contains(err.Error(), "unexpected EOF") {
					log.Error().Err(err).Msgf("parsing log entry: %s", err)
					log.Error().Msgf("Try to manually fetch the logs from %q", u.String())

					bodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to read response body: %s", err)
						errs++
						break
					}
					bodyString := string(bodyBytes)
					log.Error().Msgf("Response Body: %s", bodyString)
					
					errs++
					break
				}
					errs++
					break
				}
				log.Error().Err(err).Msgf("parsing log entry: %s", err)
				bad++
				// hmm, is this blocking
				time.Sleep(5 * time.Second) // wait a bit before retrying
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
			err = r.DB.Create(newEntries).Error
			if err != nil {
				return fmt.Errorf("inserting log entry into database: %w", err)
			}
		}

		// update timestamp & rate-limiter
		if !lastTimestamp.IsZero() {
			r.plcMutex.Lock()
			r.lastRecordTimestamp = lastTimestamp
			r.plcMutex.Unlock()

			// r.updateRateLimit(lastTimestamp)
		}

		log.Info().Msgf("%d | %d | %d | %d | %d entries. New cursor: %q", good, bad, errs, good+bad+errs, len(newEntries), cursor)
	}

	return nil
}

func (r *Runtime) AnnotatePlcLogs(start uint, batchSize int) error {
	// log := zerolog.Ctx(r.Ctx)

	var index, max atdb.ID
	index = atdb.ID(start)

	entry := atdb.PLCLogEntry{}

	err := r.DB.Model(&atdb.PLCLogEntry{}).Select("id").Order("id desc").Last(&entry).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get max id: %w", err)
	}

	max = entry.ID

	fmt.Println("Max PLC Log ID:", max)
	// log.Info().Msgf("Starting PLC log annotation backfill...")
	// log.Info().Msgf("Max PLC Log ID: %d", max)

	var good, bad, errs int

	for index < max {
		fmt.Println("Processing:", index, good, bad, errs)

		// get the next batch of PLC log entries
		var entries []atdb.PLCLogEntry
		err := r.DB.Model(&atdb.PLCLogEntry{}).
			Where("id > ?", index).
			Order("id asc").
			Limit(batchSize).
			Find(&entries).Error
		if err != nil {
			return fmt.Errorf("failed to get PLC log entries: %w", err)
		}
		if len(entries) == 0 {
			// log.Info().Msgf("No more PLC log entries to process. Backfill complete.")
			fmt.Println("No more PLC log entries to process. Backfill complete.")
			break
		}

		// Process each entry
		for _, row := range entries {
			var notes []string

			entry := atdb.PLCLogEntryToOp(row)
			// turn the entry into a PLC operation
			var op plc.Op
			switch v := row.Operation.Value.(type) {
			case plc.Op:
				op = v
			case plc.LegacyCreateOp:
				op = v.AsUnsignedOp()
			}

			_, err := syntax.ParseDID(entry.DID)
			if err != nil {
				notes = append(notes, "DID:parse")
			}

			// check handle
			if len(op.AlsoKnownAs) > 0 {
				// check all AKA
				b := 0
				for i, aka := range op.AlsoKnownAs {
					if b > 3 {
						notes = append(notes, "HDL:too-many-errs")
						break
					}
					if !strings.HasPrefix(aka, "at://") {
						notes = append(notes, fmt.Sprintf("HDL:%d:no-at", i))
						b++
						continue
					}
					handle := strings.TrimPrefix(aka, "at://")
					if handle == "" {
						notes = append(notes, fmt.Sprintf("HDL:%d:empty", i))
						b++
						continue
					}

					// this is a known bad value that showed up with PDS https://uwu
					if strings.HasPrefix(handle, "data:x") {
						notes = append(notes, fmt.Sprintf("HDL:%d:data-x", i))
						b++
						continue
					} else {
						if len(handle) > 253 {
							notes = append(notes, fmt.Sprintf("HDL:%d:length", i))
							b++
							continue
						}

						if !handleRegex.MatchString(handle) {
							notes = append(notes, fmt.Sprintf("HDL:%d:regex", i))
							b++
							continue
						}
					}
				}
			}

			// check pds
			if svc, ok := op.Services["atproto_pds"]; ok {
				pds := svc.Endpoint

				// check if pds is a known bad value
				p := sort.SearchStrings(KnownBadPDS, pds)
				if p < len(KnownBadPDS) && KnownBadPDS[p] == pds {
					notes = append(notes, "PDS:known-bad")
				}

				// check if pds is a valid URL
				pdsURL, err := url.Parse(pds)
				if err != nil {
					notes = append(notes, "PDS:parse")
				} else {
					if pdsURL.Scheme != "https" ||
						pdsURL.Path != "" ||
						pdsURL.RawQuery != "" ||
						pdsURL.Fragment != "" ||
						pdsURL.Port() != "" ||
						pdsURL.User != nil {
						notes = append(notes, "PDS:not-canonical")
					}
				}
			} else {
				notes = append(notes, "PDS:not-set")
			}

			// only try to make doc if we have no issues yet
			if len(notes) == 0 {
				doc, err := plc.MakeDoc(entry, op)
				if err != nil {
					notes = append(notes, "DOC:make-doc")
				} else {
					_, err = json.Marshal(doc)
					if err != nil {
						notes = append(notes, "DOC:marshal")
					}
				}
			}

			// TODO, try to look up account on PDS
			// or perhaps on another filter level / pass where we call describeRepo anyway

			if len(notes) > 0 {
				row.Notes = strings.Join(notes, "; ")
				bad++
			} else {
				good++
			}

			// Update the entry with notes
			err = r.DB.Model(&atdb.PLCLogEntry{}).
				Where("id = ?", row.ID).
				Updates(atdb.PLCLogEntry{
					Notes:    row.Notes,
					Filtered: len(notes),
				}).Error
			if err != nil {
				errs++
				fmt.Printf("Failed to update PLC log entry %d: %s\n", row.ID, err)
			}

			// Update the index to the last processed entry
			index = row.ID
		}

	}

	return nil
}
