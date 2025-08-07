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

var repoBackfillDescribeCmdStart uint

func init() {
	backfillCmd.AddCommand(backfillDescribeRepoCmd)
}

const backfillDescribeRepoLongHelp = `
Backfill the describeRepo for active DIDs.

flag specialization:
  --start timestamp - updates rows older than the value
`

var backfillDescribeRepoCmd = &cobra.Command{
	Use:   "describe-repo",
	Short: "Backfill the describeRepo for active DIDs",
	Long:  backfillDescribeRepoLongHelp,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		ctx, err := config.SetupLogging(ctx)
		if err != nil {
			return err
		}
		log := zerolog.Ctx(ctx).With().
			Str("module", "backfill").
			Str("method", "describe-repo").
			Logger()
		log.Info().Msgf("Starting up...")

		// create our runtime
		r, err := runtime.NewRuntime(ctx)
		if err != nil {
			log.Error().Msgf("failed to create runtime: %s", err)
			return err
		}

		err = r.BackfillDescribeRepo(backfillParallel, backfillStart)
		if err != nil {
			log.Error().Msgf("failed to backfill PLC logs: %s", err)
			return err
		}

		return nil
	},
}
