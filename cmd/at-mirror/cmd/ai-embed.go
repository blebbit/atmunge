package cmd

import (
	"fmt"

	"github.com/blebbit/at-mirror/pkg/util"
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
		uri, err := util.GetInput(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get input")
		}
		model, _ := cmd.Flags().GetString("model")
		prompt, _ := cmd.Flags().GetString("prompt")
		log.Info().Msgf("embedding post: %s", uri)

		a, err := ai.NewAI()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create AI client")
		}

		ctx := cmd.Context()
		if err := a.Embed(ctx, model, prompt, uri); err != nil {
			log.Fatal().Err(err).Msg("failed to embed")
		}

		fmt.Println("ok")
	},
}
