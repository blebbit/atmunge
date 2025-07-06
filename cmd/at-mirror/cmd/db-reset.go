package cmd

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/db"
)

func init() {
	dbCmd.AddCommand(dbResetCmd)
}

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the database",
	Long:  `Reset the database by dropping all tables and recreating them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		ctx = config.SetupLogging(ctx)
		log := zerolog.Ctx(ctx)
		log.Info().Msgf("Starting up...")

		cfg := config.GetConfig()

		// db setup
		DB, err := db.GetClient(cfg.DBUrl, ctx)
		if err != nil {
			return err
		}
		log.Info().Msgf("DB connection established")

		// db migrations (if needed)
		err = db.DropTables(DB)
		if err != nil {
			return err
		}
		err = db.MigrateModels(DB)
		if err != nil {
			return err
		}

		log.Info().Msgf("DB reset")

		return nil
	},
}
