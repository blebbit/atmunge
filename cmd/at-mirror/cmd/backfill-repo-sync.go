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
	backfillCmd.AddCommand(backfillRepoSyncCmd)
}

const backfillRepoSyncLongHelp = `
Backfill the repo sync for active DIDs.

flag specialization:
  --start timestamp - updates rows older than the value
`

var backfillRepoSyncCmd = &cobra.Command{
	Use:   "repo-sync",
	Short: "Backfill the repo sync for active DIDs",
	Long:  backfillRepoSyncLongHelp,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		ctx, err := config.SetupLogging(ctx)
		if err != nil {
			return err
		}
		log := zerolog.Ctx(ctx).With().
			Str("module", "backfill").
			Str("method", "repo-sync").
			Logger()
		log.Info().Msgf("Starting up...")

		// create our runtime
		r, err := runtime.NewRuntime(ctx)
		if err != nil {
			log.Error().Msgf("failed to create runtime: %s", err)
			return err
		}

		err = r.BackfillRepoSync(backfillParallel, backfillStart)
		if err != nil {
			log.Error().Msgf("failed to backfill repo sync: %s", err)
			return err
		}

		return nil
	},
}
