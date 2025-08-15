package db

import (
	"context"
	"database/sql"
	"fmt"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/db"
	"github.com/blebbit/at-mirror/pkg/runtime"
	appsql "github.com/blebbit/at-mirror/pkg/sql"
	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func init() {
	DBCmd.AddCommand(dbMigrateCmd)
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate [database] [args...]",
	Short: "Run database migrations",
	Long:  `Run the database migrations to update the schema.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		ctx, err := config.SetupLogging(ctx)
		if err != nil {
			return err
		}
		log := zerolog.Ctx(ctx).With().
			Str("module", "db").
			Str("method", "migrate").
			Logger()

		if len(args) < 1 {
			return fmt.Errorf("database name required")
		}
		dbName := args[0]

		// create our runtime
		r, err := runtime.NewRuntime(ctx)
		if err != nil {
			log.Error().Msgf("failed to create runtime: %s", err)
			return err
		}

		switch dbName {
		case "atm":
			// db migrations (if needed)
			err = db.MigrateModels(r.DB)
			if err != nil {
				return err
			}
			log.Info().Msgf("atm DB schema updated")
		case "acct":
			if len(args) != 2 {
				return fmt.Errorf("handle or DID is required for acct migration")
			}
			handleOrDID := args[1]

			did, _, err := r.ResolveDid(ctx, handleOrDID)
			if err != nil {
				return fmt.Errorf("failed to resolve did for %s: %w", handleOrDID, err)
			}

			dbPath := filepath.Join(r.Cfg.RepoDataDir, did, "repo.duckdb")

			dbConn, err := sql.Open("duckdb", dbPath)
			if err != nil {
				return fmt.Errorf("failed to open duckdb database at %s: %w", dbPath, err)
			}
			defer dbConn.Close()

			// db migrations (if needed)
			err = db.RunDuckDBMigrationsFromFS(dbConn, appsql.SQLFiles, "acct/migrations")
			if err != nil {
				return err
			}
			log.Info().Msgf("acct DB schema updated for %s", dbPath)
		case "network":
			log.Info().Msgf("network migration not implemented yet")
		default:
			return fmt.Errorf("unknown database: %s", dbName)
		}

		return nil
	},
}
