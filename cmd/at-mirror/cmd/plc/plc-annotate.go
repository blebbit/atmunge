package plc

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	plcAnnotateCmdStart     uint
	plcAnnotateCmdBatchSize int
)

func init() {
	PLCCmd.AddCommand(plcAnnotateCmd)
	plcAnnotateCmd.Flags().UintVar(&plcAnnotateCmdStart, "start", 0, "Start from this PLC log entry ID")
	plcAnnotateCmd.Flags().IntVar(&plcAnnotateCmdBatchSize, "batch", 1000, "Number of PLC log entries to process in one batch")
}

var plcAnnotateCmd = &cobra.Command{
	Use:   "annotate",
	Short: "Annotate the PLC logs, notably marking issues with log entries",
	Long:  "Annotate the PLC logs, notably marking issues with log entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		ctx, err := config.SetupLogging(ctx)
		if err != nil {
			return err
		}
		log := zerolog.Ctx(ctx).With().
			Str("module", "plc").
			Str("method", "annotate").
			Logger()

		// create our runtime
		r, err := runtime.NewRuntime(ctx)
		if err != nil {
			log.Error().Msgf("failed to create runtime: %s", err)
			return err
		}

		err = r.AnnotatePlcLogs(plcAnnotateCmdStart, plcAnnotateCmdBatchSize)
		if err != nil {
			log.Error().Msgf("failed to annotate PLC logs: %s", err)
			return err
		}

		return nil
	},
}
