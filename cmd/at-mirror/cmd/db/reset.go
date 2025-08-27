package db

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/blebbit/atmunge/pkg/config"
	"github.com/blebbit/atmunge/pkg/db"
	"github.com/blebbit/atmunge/pkg/runtime"
)

func init() {
	DBCmd.AddCommand(dbResetCmd)
}

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the database",
	Long:  `Reset the database by dropping all tables and recreating them.`,
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

		// db migrations (if needed)
		err = db.DropTables(r.DB)
		if err != nil {
			return err
		}
		err = db.MigrateModels(r.DB)
		if err != nil {
			return err
		}

		log.Info().Msgf("DB reset")

		return nil
	},
}
