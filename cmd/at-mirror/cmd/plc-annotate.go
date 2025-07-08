package cmd

import (
	"context"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/db"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	plcAnnotateCmdStart     uint
	plcAnnotateCmdBatchSize int
)

func init() {
	plcCmd.AddCommand(plcAnnotateCmd)
	plcAnnotateCmd.Flags().UintVar(&plcAnnotateCmdStart, "start", 0, "Start from this PLC log entry ID")
	plcAnnotateCmd.Flags().IntVar(&plcAnnotateCmdBatchSize, "batch", 1000, "Number of PLC log entries to process in one batch")
}

var plcAnnotateCmd = &cobra.Command{
	Use:   "annotate",
	Short: "Annotate the PLC logs, notably marking issues with log entries",
	Long:  "Annotate the PLC logs, notably marking issues with log entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		// ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		// defer stop()
		ctx := context.Background()

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

		err = r.AnnotatePlcLogs(plcAnnotateCmdStart, plcAnnotateCmdBatchSize)
		if err != nil {
			log.Error().Msgf("failed to annotate PLC logs: %s", err)
			return err
		}

		return nil
	},
}
