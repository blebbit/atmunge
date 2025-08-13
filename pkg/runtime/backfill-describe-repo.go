package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	atdb "github.com/blebbit/at-mirror/pkg/db"
	"github.com/rs/zerolog"
	"github.com/wandb/parallel"

	// "https://github.com/vgarvardt/gue" // alternative to parallel that is more like pg-boss

	"gorm.io/gorm/clause"
)

func (r *Runtime) BackfillDescribeRepo(par int, start string) error {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "repo-describe").Logger()

	// create a group of workers
	group := parallel.Limited(r.Ctx, par)

	// get total count of PdsRepo entries to process for progress reporting
	count, err := r.countRemainingToProcess("account_infos", start, "updated_at")
	if err != nil {
		return fmt.Errorf("failed to count repo describes: %w", err)
	}

	for {

		dids, err := r.getRandomSetToProcess("account_infos", start, "updated_at", 1000)
		if err != nil {
			return fmt.Errorf("failed to get random repo describes: %w", err)
		}

		// if no more entries, we are done
		if len(dids) == 0 {
			log.Info().Msgf("No PdsRepo entries found to fetch, exiting")
			break
		}

		var total atomic.Int64

		for index := 0; index < len(dids); index++ {
			group.Go(func(ctx context.Context) {
				err := r.processRepoDescribe(dids[index])
				if err != nil {
					log.Error().Err(err).Msgf("Failed to process repo %s", dids[index])
					return
				}

				total.Add(1)
			})
		}

		log.Info().Msgf("Processing %d/%d repos", total.Load(), count)
	}

	return nil
}

func (r *Runtime) processRepoDescribe(did string) error {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "repo-describe").Logger()

	// get the next PdsRepo entry to process
	var row atdb.PdsRepo
	err := r.DB.Model(&atdb.PdsRepo{}).
		Where("did = ?", did).
		First(&row).Error
	if err != nil {
		return fmt.Errorf("failed to get PdsRepo entry: %w", err)
	}

	// per PDS rate limiters, blocks until some rate limit is available
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.describeRepo?repo=%s", row.PDS, row.DID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %w", url, err)
	}
	resp, err := r.Proxy.Do(req)
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
