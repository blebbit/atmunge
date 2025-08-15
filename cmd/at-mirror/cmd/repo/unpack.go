package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/repo"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/bluesky-social/indigo/atproto/data"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/ipfs/go-cid"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var repoUnpackCmd = &cobra.Command{
	Use:   "unpack [acct]",
	Short: "Unpack records from a CAR file for an account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		ctx, err := config.SetupLogging(ctx)
		if err != nil {
			return err
		}
		log := zerolog.Ctx(ctx).With().
			Str("module", "repo").
			Str("method", "unpack").
			Logger()

		// create our runtime
		r, err := runtime.NewRuntime(ctx)
		if err != nil {
			log.Error().Msgf("failed to create runtime: %s", err)
			return err
		}

		handleOrDID := args[0]
		did, _, err := r.ResolveDid(ctx, handleOrDID)
		if err != nil {
			return fmt.Errorf("failed to resolve did for %s: %w", handleOrDID, err)
		}

		carFile := filepath.Join(r.Cfg.RepoDataDir, did, "repo.car")
		outputDir := filepath.Join(r.Cfg.RepoDataDir, did, "unpacked")

		f, err := os.Open(carFile)
		if err != nil {
			return fmt.Errorf("failed to open car file %s: %w", carFile, err)
		}
		defer f.Close()

		repo, err := repo.ReadRepoFromCar(f)
		if err != nil {
			return fmt.Errorf("failed to read repo from car: %w", err)
		}

		log.Info().Msgf("Unpacking records from %s to %s", carFile, outputDir)

		err = repo.MST.Walk(func(k []byte, v cid.Cid) error {
			col, rkey, err := syntax.ParseRepoPath(string(k))
			if err != nil {
				return err
			}
			recBytes, _, err := repo.GetRecordBytes(ctx, col, rkey)
			if err != nil {
				return err
			}

			rec, err := data.UnmarshalCBOR(recBytes)
			if err != nil {
				return err
			}

			recPath := filepath.Join(outputDir, string(k)+".json")
			log.Debug().Msgf("Unpacking %s", recPath)
			err = os.MkdirAll(filepath.Dir(recPath), os.ModePerm)
			if err != nil {
				return err
			}
			recJSON, err := json.MarshalIndent(rec, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(recPath, recJSON, 0666); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk repo: %w", err)
		}

		log.Info().Msgf("Successfully unpacked records from %s to %s", carFile, outputDir)
		return nil
	},
}

func init() {
	// no-op
}
