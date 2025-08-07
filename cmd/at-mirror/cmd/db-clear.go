package cmd

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/db"
	"github.com/blebbit/at-mirror/pkg/runtime"
)

func init() {
	dbCmd.AddCommand(dbClearCmd)
}

var dbClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the database",
	Long:  `Clear the database by dropping all tables.`,
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

		// create our runtime
		r, err := runtime.NewRuntime(ctx)
		if err != nil {
			log.Error().Msgf("failed to create runtime: %s", err)
			return err
		}

		// db migrations (if needed)
		err = db.ClearTables(r.DB)
		if err != nil {
			return err
		}
		log.Info().Msgf("DB cleared")

		return nil
	},
}
