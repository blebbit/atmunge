package acct

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blebbit/at-mirror/pkg/repo"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog/log"
)

func Sync(rt *runtime.Runtime, handleOrDID string, phases []string) error {
	if len(phases) == 0 {
		phases = []string{"car", "duckdb", "blobs"}
	}
	log.Info().Msgf("Syncing account: %s, with phases: %v", handleOrDID, phases)
	ctx := context.Background()

	// 1. Resolve handle/DID and get PDS
	did, pds, err := rt.ResolveDid(ctx, handleOrDID)
	if err != nil {
		return fmt.Errorf("failed to resolve %s: %w", handleOrDID, err)
	}
	log.Info().Msgf("Resolved %s to %s on PDS %s", handleOrDID, did, pds)

	repoDir := filepath.Join(rt.Cfg.RepoDataDir, did)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return fmt.Errorf("failed to create repo directory for %s: %w", did, err)
	}

	carPath := filepath.Join(repoDir, "repo.car")
	duckdbPath := filepath.Join(repoDir, "repo.duckdb")

	for _, phase := range phases {
		switch phase {
		case "car":
			log.Info().Msgf("Syncing CAR file to %s", carPath)
			blockstore, since, err := repo.LoadLocalCar(carPath)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("failed to load local car: %w", err)
			}

			carData, err := repo.GetRepo(rt.Ctx, rt.Proxy, pds, did, since)
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
		case "duckdb":
			log.Info().Msgf("Converting CAR to DuckDB at %s", duckdbPath)
			if err := repo.CarToDuckDB(rt.Ctx, carPath, duckdbPath); err != nil {
				return fmt.Errorf("failed to convert car to duckdb: %w", err)
			}
			log.Info().Msg("DuckDB conversion complete")
		case "blobs":
			log.Info().Msg("Syncing blobs")
			repo.SyncBlobs(pds, did, rt.Cfg.RepoDataDir)
			log.Info().Msg("Blob sync complete")
		default:
			return fmt.Errorf("unknown phase: %s", phase)
		}
	}

	log.Info().Msgf("Successfully synced account %s", handleOrDID)
	return nil
}
