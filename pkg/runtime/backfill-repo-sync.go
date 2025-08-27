package runtime

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/wandb/parallel"
	"gorm.io/gorm/clause"

	atdb "github.com/blebbit/atmunge/pkg/db"
	"github.com/blebbit/atmunge/pkg/repo"
)

// we should build a channel into this
// some downloads take longer for various reasons
// we should keep the belt full (like factorio)

// XXX the counts are off too, by 20% it seems like, not sure what's up with that...
//     I suspect the parallel module, let's try writing our own goroutines instead

// XXX we also need to handle account status responses here
//     between the time we filter and then download CARs
//     the account may have been deleted or taken down

func (r *Runtime) BackfillRepoSync(par int, start string) error {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "repo-sync").Logger()

	// create a group of workers
	group := parallel.Limited(r.Ctx, par)

	// Can we get the latest revision from the PDS?
	// If so, we can skip accounts without fetching CAR to see if there are updates
	// Since we cascade the backfill process, one of the prior processes or tables may have the latest revision

	// XXX this process probably does not account for when...
	// 1. we run an update AND there are new accounts
	// 2. right now we split the logic
	// we really want to have it be seamless
	// for now we could run it twice, once in each mode, this should be equivalent

	// get total count of AccountRepo entries to process for progress reporting
	count, err := r.countRemainingToProcess("account_repos", start, "updated_at")
	if err != nil {
		return fmt.Errorf("failed to count repo syncs: %w", err)
	}

	// TODO, have both per batch and run stats
	var total atomic.Int64
	var totalErr atomic.Int64

	for {
		var batch atomic.Int64
		var batchErr atomic.Int64

		dids, err := r.getRandomSetToProcess("account_repos", start, "updated_at", 50)
		if err != nil {
			return fmt.Errorf("failed to get random repo syncs: %w", err)
		}

		// if no more entries, we are done
		if len(dids) == 0 {
			log.Info().Msgf("No AccountRepo entries found to sync, exiting")
			break
		}

		for index := 0; index < len(dids); index++ {
			group.Go(func(ctx context.Context) {
				err := r.processRepoSync(dids[index])
				total.Add(1)
				batch.Add(1)

				if err != nil {
					totalErr.Add(1)
					batchErr.Add(1)
					log.Error().Err(err).Msgf("Failed to process repo sync for %s: %v", dids[index], err)
					return
				}
			})
		}

		log.Info().Msgf("Processing %d/%d repos. Batch: %d/%d Error: %d/%d",
			total.Load(), count,
			batch.Load(), len(dids),
			batchErr.Load(), totalErr.Load(),
		)
	}

	return nil
}

func (r *Runtime) processRepoSync(did string) error {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "repo-sync").Logger()

	var account atdb.AccountInfo
	err := r.DB.Model(&atdb.AccountInfo{}).
		Where("did = ?", did).
		First(&account).Error
	if err != nil {
		return fmt.Errorf("failed to get AccountInfo entry: %w", err)
	}

	pdsHost := account.PDS
	if pdsHost == "" {
		return fmt.Errorf("account %s has no PDS host", did)
	}

	// prepare to write
	repoDir := filepath.Join(r.Cfg.RepoDataDir, did)
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		return fmt.Errorf("failed to create data directory %s: %w", repoDir, err)
	}
	localCarFile := filepath.Join(repoDir, "repo.car")
	// load local CAR from disk
	blockstoreMem, sinceTID, err := repo.LoadLocalCar(localCarFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to load local CAR for %s: %w", did, err)
		}
		// if the file doesn't exist, we can continue, it's a new repo
	}

	// get updated CAR data from PDS
	updateCarData, err := repo.GetRepo(r.Ctx, r.Proxy, pdsHost, did, sinceTID)
	if err != nil {
		return fmt.Errorf("failed to fetch repo data for %s: %w", did, err)
	}

	val := atdb.AccountRepo{
		DID: did,
	}

	onConflict := clause.OnConflict{
		Columns:   []clause.Column{{Name: "did"}},
		DoUpdates: clause.AssignmentColumns([]string{"rev", "last_changed", "updated_at"}),
	}

	// within the following code until we write to the database
	// update time fields so we can differentiate between
	// - last_changed (logically)
	// - updated_at (last time we tried this backfill session)

	// if we have updates, merge them
	if len(updateCarData) > 0 {

		newRootCid, newestRev, _ /*newBlocks*/, err := repo.MergeUpdate(blockstoreMem, updateCarData)
		if err != nil {
			return fmt.Errorf("failed to merge update for %s: %w", did, err)
		}

		// this second check is a possible edge case or shouldn't happen logically, let's be defensively programming anyhow
		if newestRev != "" {
			val.LastChanged = time.Now()
			val.Rev = newestRev
			if err := repo.WriteCar(localCarFile, newRootCid, blockstoreMem); err != nil {
				return fmt.Errorf("failed to write CAR file for %s: %w", did, err)
			}

			// TODO, these should be controlled by flags
			//   right now, it significantly increase the disk footprint
			// also write to sqlite & duckdb
			// duckdbPath := filepath.Join(repoDir, "repo.duckdb")
			// if err := repo.SaveNewRecordsToDuckDB(r.Ctx, blockstoreMem, newBlocks, newRootCid, duckdbPath); err != nil {
			// 	log.Error().Err(err).Msgf("failed to save new records to DuckDB for %s", did)
			// 	// deciding not to return error here, as the primary sync succeeded
			// }
			// sqlitePath := filepath.Join(repoDir, "repo.sqlite")
			// if err := repo.SaveNewRecordsToSQLite(r.Ctx, blockstoreMem, newBlocks, newRootCid, sqlitePath); err != nil {
			// 	log.Error().Err(err).Msgf("failed to save new records to SQLite for %s", did)
			// 	// deciding not to return error here, as the primary sync succeeded
			// }
		}
	} else {
		// we want to note that we have checked this record, but there are no other changes
		// Go has default values that could wipe out our existing values
		onConflict.DoUpdates = clause.AssignmentColumns([]string{"updated_at"})
	}

	err = r.DB.
		Model(&atdb.AccountRepo{}).
		Clauses(onConflict).
		Create(&val).Error
	if err != nil {
		log.Error().Err(err).Msgf("failed to create AccountRepo entry for %s", did)
		return err
	}

	return nil
}
