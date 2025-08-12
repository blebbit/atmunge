package acct

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blebbit/at-mirror/pkg/repo"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog/log"
)

func Sync(rt *runtime.Runtime, handleOrDID string, phase string) error {
	log.Info().Msgf("Syncing account: %s, starting from phase: %q", handleOrDID, phase)
	ctx := context.Background()

	// 1. Resolve handle/DID and get PDS
	did, pds, err := rt.ResolveDid(ctx, handleOrDID)
	if err != nil {
		return fmt.Errorf("failed to resolve %s: %w", handleOrDID, err)
	}
	log.Info().Msgf("Resolved %s to %s on PDS %s", handleOrDID, did, pds)

	carPath := filepath.Join(rt.Cfg.RepoDataDir, did+".car")
	duckdbPath := filepath.Join(rt.Cfg.RepoDataDir, did+".duckdb")

	switch phase {
	case "":
		fallthrough
	case "car":
		log.Info().Msgf("Syncing CAR file to %s", carPath)
		blockstore, since, err := repo.LoadLocalCar(carPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to load local car: %w", err)
		}

		carData, err := repo.GetRepo(pds, did, since)
		if err != nil {
			return fmt.Errorf("failed to get repo: %w", err)
		}

		if len(carData) > 0 {
			newRoot, _, _, err := repo.MergeUpdate(blockstore, carData)
			if err != nil {
				return fmt.Errorf("failed to merge update: %w", err)
			}
			if err := repo.WriteCar(carPath, newRoot, blockstore); err != nil {
				return fmt.Errorf("failed to write car: %w", err)
			}
			log.Info().Msg("CAR file updated")
		} else {
			log.Info().Msg("CAR file is up to date")
		}
		fallthrough
	case "duckdb":
		log.Info().Msgf("Converting CAR to DuckDB at %s", duckdbPath)
		if err := repo.CarToDuckDB(carPath, duckdbPath); err != nil {
			return fmt.Errorf("failed to convert car to duckdb: %w", err)
		}
		log.Info().Msg("DuckDB conversion complete")
		fallthrough
	case "blobs":
		log.Info().Msg("Syncing blobs")
		repo.SyncBlobs(pds, did, rt.Cfg.RepoDataDir)
		log.Info().Msg("Blob sync complete")
	default:
		return fmt.Errorf("unknown phase: %s", phase)
	}

	log.Info().Msgf("Successfully synced account %s", handleOrDID)
	return nil
}
