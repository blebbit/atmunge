package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/blebbit/at-mirror/pkg/ai"
)

func init() {
	aiCmd.AddCommand(aiEmbedCmd)
}

var aiEmbedCmd = &cobra.Command{
	Use:   "embed <at-uri>",
	Short: "Embed a post",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}
		uri := args[0]
		log.Info().Msgf("embedding post: %s", uri)

		a, err := ai.NewAI()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create AI client")
		}

		ctx := cmd.Context()
		if err := a.Embed(ctx, uri); err != nil {
			log.Fatal().Err(err).Msg("failed to embed")
		}

		fmt.Println("ok")
	},
}
