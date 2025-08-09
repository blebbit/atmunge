package runtime

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/rs/zerolog"
	"github.com/wandb/parallel"
	"gorm.io/gorm/clause"

	atdb "github.com/blebbit/at-mirror/pkg/db"
	"github.com/blebbit/at-mirror/pkg/repo"
)

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

	for {
		ids, err := r.getRandomSetToProcess("account_repos", start, "updated_at", 1000)
		if err != nil {
			return fmt.Errorf("failed to get random repo syncs: %w", err)
		}

		// if no more entries, we are done
		if len(ids) == 0 {
			log.Info().Msgf("No AccountRepo entries found to sync, exiting")
			break
		}

		var total atomic.Int64

		for index := 0; index < len(ids); index++ {
			group.Go(func(ctx context.Context) {
				err := r.processRepoSync(ids[index])
				if err != nil {
					log.Error().Err(err).Msgf("Failed to process repo sync for %s", ids[index])
					return
				}

				total.Add(1)
			})
		}

		log.Info().Msgf("Processing %d/%d repos", total.Load(), count)
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

	// per PDS rate limiters, blocks until some rate limit is available
	r.limitTaker(pdsHost)

	// prepare to write
	dataDir := r.Cfg.RepoDataDir
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("failed to create data directory %s: %w", dataDir, err)
	}
	localCarFile := fmt.Sprintf("%s/%s.car", dataDir, did)

	// load local CAR from disk
	blockstoreMem, sinceTID, err := repo.LoadLocalCar(localCarFile)
	if err != nil {
		return fmt.Errorf("failed to load local CAR for %s: %w", did, err)
	}

	// get updated CAR data from PDS
	updateCarData, err := repo.GetRepo(pdsHost, did, sinceTID)
	if err != nil {
		return fmt.Errorf("failed to fetch repo data for %s: %w", did, err)
	}

	var newRootCid cid.Cid
	var newestRev string

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
		newRootCid, newestRev, err = repo.MergeUpdate(blockstoreMem, updateCarData)
		if err != nil {
			return fmt.Errorf("failed to merge update for %s: %w", did, err)
		}

		// this second check is a possible edge case or shouldn't happen logically, let's be defensively programming anyhow
		if newestRev == "" {
			val.LastChanged = time.Now()
			val.Rev = newestRev
			if err := repo.WriteCar(localCarFile, newRootCid, blockstoreMem); err != nil {
				return fmt.Errorf("failed to write CAR file for %s: %w", did, err)
			}
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
