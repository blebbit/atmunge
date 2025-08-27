package ai

import (
	"fmt"

	"github.com/blebbit/atmunge/pkg/ai"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	AICmd.AddCommand(aiReplyCmd)
}

var aiReplyCmd = &cobra.Command{
	Use:   "reply <prompt>",
	Short: "Generate a reply to a string",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}
		a, err := ai.NewAI()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create AI client")
		}

		userPrompt, err := a.ResolveInput(args[0])
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get input")
		}
		model, _ := cmd.Flags().GetString("model")
		systemPrompt, _ := cmd.Flags().GetString("prompt")

		ctx := cmd.Context()
		resp, err := a.Reply(ctx, model, systemPrompt, userPrompt)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to reply")
		}

		fmt.Println(resp)
	},
}
