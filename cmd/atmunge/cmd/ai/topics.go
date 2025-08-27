package ai

import (
	"fmt"

	"github.com/blebbit/atmunge/pkg/ai"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	AICmd.AddCommand(aiTopicsCmd)
}

var aiTopicsCmd = &cobra.Command{
	Use:   "topics <at-uri>",
	Short: "Get topics for a post",
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
		log.Info().Msgf("getting topics for post: %s", uri)

		ctx := cmd.Context()
		if err := a.Topics(ctx, model, prompt, uri); err != nil {
			log.Fatal().Err(err).Msg("failed to get topics")
		}

		fmt.Println("ok")
	},
}
