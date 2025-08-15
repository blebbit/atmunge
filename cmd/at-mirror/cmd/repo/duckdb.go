package repo

import (
	"context"
	"fmt"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/repo"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var repoDuckDBCmd = &cobra.Command{
	Use:   "duckdb [acct]",
	Short: "Convert a CAR file to a DuckDB database for an account",
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
			Str("method", "duckdb").
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
		dbPath := filepath.Join(r.Cfg.RepoDataDir, did, "repo.duckdb")

		log.Info().Msgf("Converting %s to %s", carFile, dbPath)
		if err := repo.CarToDuckDB(carFile, dbPath); err != nil {
			return fmt.Errorf("failed to convert CAR to DuckDB: %w", err)
		}
		log.Info().Msgf("Successfully converted %s to %s", carFile, dbPath)
		return nil
	},
}

func init() {
	// no-op
}
