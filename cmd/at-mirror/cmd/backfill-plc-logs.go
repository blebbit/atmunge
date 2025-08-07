package cmd

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func init() {
	backfillCmd.AddCommand(backfillPlcLogsCmd)
}

var backfillPlcLogsCmd = &cobra.Command{
	Use:   "plc-logs",
	Short: "Backfill the PLC logs",
	Long:  `Synchronize the PLC logs into the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		ctx, err := config.SetupLogging(ctx)
		if err != nil {
			return err
		}
		log := zerolog.Ctx(ctx).With().
			Str("module", "backfill").
			Str("method", "plc-logs").
			Logger()
		log.Info().Msgf("Starting up...")

		// create our runtime
		r, err := runtime.NewRuntime(ctx)
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
