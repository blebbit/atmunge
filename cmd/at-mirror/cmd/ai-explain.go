package cmd

import (
	"fmt"

	"github.com/blebbit/at-mirror/pkg/ai"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	aiCmd.AddCommand(aiExplainCmd)
}

var aiExplainCmd = &cobra.Command{
	Use:   "explain <at-uri>",
	Short: "Explain a post",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}
		a, err := ai.NewAI()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create AI client")
		}

		uri, err := a.ResolveInput(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get input")
		}
		model, _ := cmd.Flags().GetString("model")
		prompt, _ := cmd.Flags().GetString("prompt")
		log.Info().Msgf("explaining post: %s", uri)

		ctx := cmd.Context()
		if err := a.Explain(ctx, model, prompt, uri); err != nil {
			log.Fatal().Err(err).Msg("failed to explain")
		}

		fmt.Println("ok")
	},
}
