package db

import (
	"context"
	"database/sql"
	"fmt"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/blebbit/atmunge/pkg/config"
	"github.com/blebbit/atmunge/pkg/db"
	"github.com/blebbit/atmunge/pkg/repo"
	"github.com/blebbit/atmunge/pkg/runtime"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var tablesToClear []string

func init() {
	DBCmd.AddCommand(dbClearCmd)
	dbClearCmd.Flags().StringSliceVarP(&tablesToClear, "tables", "t", []string{}, "Tables to clear (optional, clears all if not set)")
}

var dbClearCmd = &cobra.Command{
	Use:   "clear [database] [args...]",
	Short: "Clear tables from a database",
	Long:  `Clear tables from the main app database or a specific account's database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		ctx, err := config.SetupLogging(ctx)
		if err != nil {
			return err
		}
		log := zerolog.Ctx(ctx).With().
			Str("module", "db").
			Str("method", "clear").
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
		case "app":
			err = db.ClearTables(r.DB, tablesToClear)
			if err != nil {
				return err
			}
			log.Info().Msg("app DB tables cleared")
		case "acct":
			if len(args) != 2 {
				return fmt.Errorf("handle or DID is required for acct clear")
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

			err = repo.ClearDuckDBTables(dbConn, tablesToClear)
			if err != nil {
				return err
			}
			log.Info().Msgf("acct DB tables cleared for %s", dbPath)
		default:
			return fmt.Errorf("unknown database: %s", dbName)
		}

		if len(tablesToClear) == 0 {
			log.Info().Msgf("all tables cleared")
		} else {
			log.Info().Msgf("tables cleared: %v", tablesToClear)
		}

		return nil
	},
}
