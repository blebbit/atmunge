package cmd

import (
	"context"

	"github.com/blebbit/atmunge/pkg/config"
	"github.com/blebbit/atmunge/pkg/firehose"
	"github.com/blebbit/atmunge/pkg/runtime"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var firehoseCmd = &cobra.Command{
	Use:   "firehose",
	Short: "Connect to the firehose and sync commits",
	Long:  `Connect to the firehose and sync commits`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		ctx, err := config.SetupLogging(ctx)
		if err != nil {
			return err
		}
		log := zerolog.Ctx(ctx)

		r, err := runtime.NewRuntime(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create runtime")
		}

		fc, err := firehose.NewFirehoseClient(r)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create firehose client")
		}

		return fc.ConnectAndRead(ctx)
	},
}
