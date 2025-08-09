package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"

	atdb "github.com/blebbit/at-mirror/pkg/db"
	"github.com/rs/zerolog"
	"github.com/wandb/parallel"

	// "https://github.com/vgarvardt/gue" // alternative to parallel that is more like pg-boss

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Runtime) BackfillGetRepo(par int, start string) error {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "get-repo").Logger()

	// create a group of workers
	group := parallel.Limited(r.Ctx, par)

	// get total count of PdsRepo entries to process for progress reporting
	count, err := r.countRemainingToProcess("account_repos", start, "updated_at")
	if err != nil {
		return fmt.Errorf("failed to count repo describes: %w", err)
	}

	for {

		dids, err := r.getRandomSetToProcess("account_repos", start, "updated_at", 1000)
		if err != nil {
			return fmt.Errorf("failed to get random repo describes: %w", err)
		}

		// if no more entries, we are done
		if len(dids) == 0 {
			log.Info().Msgf("No PdsRepo entries found to fetch, exiting")
			break
		}

		var total atomic.Int64

		for index := range len(dids) {
			group.Go(func(ctx context.Context) {
				err := r.processGetRepo(dids[index])
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

func (r *Runtime) processGetRepo(did string) error {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "get-repo").Logger()

	// get the next PdsRepo entry to process
	var row atdb.PdsRepo
	err := r.DB.Model(&atdb.PdsRepo{}).
		Where("did = ?", did).
		First(&row).Error
	if err != nil {
		return fmt.Errorf("failed to get PdsRepo entry: %w", err)
	}

	// look for an existing AccountRepo entry
	var existing atdb.AccountRepo
	err = r.DB.Model(&atdb.AccountRepo{}).
		Where("did = ?", row.DID).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// no existing entry found, create a new one
			existing = atdb.AccountRepo{
				DID: row.DID,
			}
		} else {
			return fmt.Errorf("failed to get AccountRepo entry: %w", err)
		}
	}

	// per PDS rate limiters, blocks until some rate limit is available
	r.limitTaker(row.PDS)

	// build the URL to fetch the repo
	url := fmt.Sprintf("%s/xrpc/com.atproto.sync.getRepo?did=%s", row.PDS, row.DID)
	if row.Rev != "" {
		url += fmt.Sprintf("&since=%s", row.Rev)
	}

	// make the call
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).Msgf("failed to get repos from %s", url)
		return err
	}
	defer resp.Body.Close()

	// TODO, create the output file and directly write the response to it
	fn := fmt.Sprintf("%s/%s.car", r.Cfg.RepoDataDir, row.DID)

	f, err := os.Create(fn)
	if err != nil {
		log.Error().Err(err).Msgf("failed to create output file %s", fn)
		return err
	}
	defer f.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 429 {
			// TODO, wait some time and retry
			return fmt.Errorf("Rate limit exceeded for %s, skipping...", url)
		}

		// TODO, should we log the response body? (based on verbosity level)
		// b, err := io.ReadAll(resp.Body)
		// if err != nil {
		// 	log.Error().Err(err).Msgf("failed to read response body from %s", url)
		// 	return err
		// }

		return fmt.Errorf("bad status code from %s: %d", url, resp.StatusCode)
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		log.Error().Err(err).Msgf("failed to write response body to %s", fn)
		return err
	}

	val := atdb.AccountRepo{
		DID: row.DID,
		Rev: "tbd",
	}

	err = r.DB.
		Model(&atdb.AccountRepo{}).
		Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "did"}},
				DoUpdates: clause.AssignmentColumns([]string{"rev"}),
			},
		).
		Create(&val).Error
	if err != nil {
		log.Error().Err(err).Msgf("failed to create AccountRepo entry for %s", row.DID)
		return err
	}

	return nil
}
