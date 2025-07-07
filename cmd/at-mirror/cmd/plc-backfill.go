package cmd

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/db"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func init() {
	plcCmd.AddCommand(plcBackfillCmd)
}

var plcBackfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "Backfill the PLC logs",
	Long:  `Synchronize the PLC logs into the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		ctx = config.SetupLogging(ctx)
		log := zerolog.Ctx(ctx).With().Str("module", "plc").Logger()
		log.Info().Msgf("Starting up...")

		cfg := config.GetConfig()

		// db setup
		DB, err := db.GetClient(cfg.DBUrl, ctx)
		if err != nil {
			return err
		}
		log.Info().Msgf("DB connection established")

		// create our runtime
		r, err := runtime.NewRuntime(ctx, DB)
		if err != nil {
			log.Error().Msgf("failed to create runtime: %s", err)
			return err
		}

		err = r.BackfillPlcLogs()
		if err != nil {
			log.Error().Msgf("failed to backfill PLC logs: %s", err)
			return err
		}

		return nil
	},
}
