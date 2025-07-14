package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"

	atdb "github.com/blebbit/at-mirror/pkg/db"
	"github.com/rs/zerolog"
	"github.com/wandb/parallel"

	// "https://github.com/vgarvardt/gue" // alternative to parallel that is more like pg-boss
	"go.uber.org/ratelimit"
	"gorm.io/gorm/clause"
)

// per DSP rate limiters
var limiters sync.Map

func (r *Runtime) RepoBackfillDescribe(start, end, par int, retry_errors bool) error {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "repo-describe").Logger()

	// fetch all PdsRepo entries that have no corresponding AccountInfo entry
	// TODO, add an updated_at option so we can refresh on cron or similar
	var ids []string
	log.Info().Msgf("Gathering entries to backfill")
	err := r.DB.Model(&atdb.PdsRepo{}).
		Where("active = true").
		Where("NOT EXISTS (SELECT 1 FROM account_infos WHERE account_infos.did = pds_repos.did)").
		Pluck("id", &ids).Error
	if err != nil {
		return fmt.Errorf("failed to fetch PdsRepo entries: %w", err)
	}
	log.Info().Msgf("Found %d PdsRepo entries to fetch", len(ids))
	if len(ids) == 0 {
		log.Info().Msgf("No PdsRepo entries found to fetch, exiting")
		return nil
	}

	// Shuffle the slice (pds_repos is ordered by pds->did)
	// so we can spread the parallel requests across PDSes
	// and not get rate limited by a single PDS in order
	for i := range ids {
		j := i + int(uint32(rand.Int63())%(uint32(len(ids))-uint32(i)))
		ids[i], ids[j] = ids[j], ids[i]
	}

	// create a group of workers
	group := parallel.Limited(r.Ctx, par)

	// queue up work
	if end < 0 || end > len(ids) {
		end = len(ids)
	}
	for index := atdb.ID(start); int(index) < end; index++ {
		group.Go(func(ctx context.Context) {
			err := r.processRepoDescribe(ids[index])
			if err != nil {
				log.Error().Err(err).Msgf("Failed to process repo %s", ids[index])
				return
			}
			// <-time.After(time.Second)

			if index%1000 == 0 {
				log.Info().Msgf("Processing %d/%d repos", index, end)
			}
		})
	}
	return nil
}

func (r *Runtime) processRepoDescribe(id string) error {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "repo-describe").Logger()

	// get the next PdsRepo entry to process
	var row atdb.PdsRepo
	err := r.DB.Model(&atdb.PdsRepo{}).
		Where("id = ?", id).
		First(&row).Error
	if err != nil {
		return fmt.Errorf("failed to get PdsRepo entry: %w", err)
	}

	// per PDS rate limiters
	limiter, ok := limiters.Load(row.PDS)
	if !ok {
		// create a new rate limiter for this PDS
		limiter = ratelimit.New(9) // 10 requests per second is at the limit the PDS defaults to (3000;300w)
		limiter, _ = limiters.LoadOrStore(row.PDS, limiter)
	}
	limiter.(ratelimit.Limiter).Take() // wait for the rate limit

	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.describeRepo?repo=%s", row.PDS, row.DID)
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).Msgf("failed to get repos from %s", url)
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msgf("failed to read response body from %s", url)
		return err
	}

	// log.Info().Msgf("Resp (%d) %s: %s", resp.StatusCode, url, string(b))

	if resp.StatusCode != http.StatusOK {

		if resp.StatusCode == http.StatusNotFound {
			err = r.DB.Model(&atdb.PdsRepo{}).
				Where("id = ?", row.ID).
				Updates(atdb.PdsRepo{
					Active: false,
					Status: "notfound",
				}).Error
			if err != nil {
				return fmt.Errorf("Failed to update PdsRepo entry %d: %s\n", row.ID, err)
			}
		}

		if resp.StatusCode == http.StatusBadRequest {
			var data map[string]any
			err := json.Unmarshal(b, &data)
			if err != nil {
				return fmt.Errorf("failed to unmarshal error response from %s: %w\n", url, err)
			}

			if msg, ok := data["error"].(string); ok {
				switch msg {
				case "RepoTakendown":
					err = r.DB.Model(&atdb.PdsRepo{}).
						Where("id = ?", row.ID).
						Updates(atdb.PdsRepo{
							Active: false,
							Status: "takendown",
						}).Error
					if err != nil {
						return fmt.Errorf("Failed to update PdsRepo entry %d: %s\n", row.ID, err)
					}

				case "RepoDeactivated":
					err = r.DB.Model(&atdb.PdsRepo{}).
						Where("id = ?", row.ID).
						Updates(atdb.PdsRepo{
							Active: false,
							Status: "deactivated",
						}).Error
					if err != nil {
						return fmt.Errorf("Failed to update PdsRepo entry %d: %s\n", row.ID, err)
					}

				case "NotFound", "RepoNotFound":
					err = r.DB.Model(&atdb.PdsRepo{}).
						Where("id = ?", row.ID).
						Updates(atdb.PdsRepo{
							Active: false,
							Status: "notfound",
						}).Error
					if err != nil {
						return fmt.Errorf("Failed to update PdsRepo entry %d: %s", row.ID, err)
					}

				default:
					return fmt.Errorf("Unhandled StatusBadRequest error from %s: %s %v", url, msg, resp.StatusCode)
				}
				return nil
			} else {
				return fmt.Errorf("Unhandled StatusBadRequest no error in body %s: %s", url, string(b))

			}
		}

		if resp.StatusCode == 429 {
			// TODO, wait some time and retry
			return fmt.Errorf("Rate limit exceeded for %s, skipping...", url)
		}

		if resp.StatusCode >= 500 {
			err = r.DB.Model(&atdb.PdsRepo{}).
				Where("id = ?", row.ID).
				Updates(atdb.PdsRepo{
					Active: false,
					Status: fmt.Sprintf("server_error_%d", resp.StatusCode),
				}).Error
			if err != nil {
				return fmt.Errorf("Failed to update PdsRepo entry %d: %s", row.ID, err)
			}
			return nil
		}

		// You might want to read the body here to get more error details
		return fmt.Errorf("unhandled bad status code from %s: %d\n", url, resp.StatusCode)
	}

	val := atdb.AccountInfo{
		DID:      row.DID,
		PDS:      row.PDS,
		Describe: b,
	}

	err = r.DB.
		Model(&atdb.AccountInfo{}).
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "did"}},
				DoUpdates: clause.AssignmentColumns([]string{"pds", "describe"}),
			},
		).
		Create(&val).Error
	if err != nil {
		log.Error().Err(err).Msgf("failed to create AccountInfo entry for %s", row.DID)
		return err
	}
	// ensure record is active
	err = r.DB.Model(&atdb.PdsRepo{}).
		Where("id = ?", row.ID).
		Updates(atdb.PdsRepo{
			Active: true,
			Status: "",
		}).Error
	if err != nil {
		return fmt.Errorf("Failed to update PdsRepo entry %d: %s", row.ID, err)
	}

	return nil
}
