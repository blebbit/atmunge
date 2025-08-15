package db

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/db"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func init() {
	DBCmd.AddCommand(dbMigrateCmd)
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate [database]",
	Short: "Run database migrations",
	Long:  `Run the database migrations to update the schema.`,
	Args:  cobra.ExactArgs(1),
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

		// create our runtime
		r, err := runtime.NewRuntime(ctx)
		if err != nil {
			log.Error().Msgf("failed to create runtime: %s", err)
			return err
		}

		dbName := args[0]
		switch dbName {
		case "atm":
			// db migrations (if needed)
			err = db.MigrateModels(r.DB)
			if err != nil {
				return err
			}
			log.Info().Msgf("atm DB schema updated")
		case "acct":
			// get the path to the migrations directory
			// HACK: this is not a good way to do this
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current working directory: %w", err)
			}
			migrationsPath := filepath.Join(cwd, "pkg", "sql", "acct", "migrations")

			sqlDB, err := r.DB.DB()
			if err != nil {
				return fmt.Errorf("failed to get sql.DB: %w", err)
			}
			// db migrations (if needed)
			err = db.RunDuckDBMigrations(sqlDB, migrationsPath)
			if err != nil {
				return err
			}
			log.Info().Msgf("acct DB schema updated")
		case "network":
			log.Info().Msgf("network migration not implemented yet")
		default:
			return fmt.Errorf("unknown database: %s", dbName)
		}

		return nil
	},
}
