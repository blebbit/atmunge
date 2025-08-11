package cmd

import (
	"fmt"

	"github.com/blebbit/at-mirror/pkg/ai"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	aiCmd.AddCommand(aiChatCmd)
	aiChatCmd.Flags().String("model", "llama3", "The model to use for chatting")
}

var aiChatCmd = &cobra.Command{
	Use:   "chat <prompt>",
	Short: "Chat with a model",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}
		prompt := args[0]
		model, _ := cmd.Flags().GetString("model")

		a, err := ai.NewAI()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create AI client")
		}

		ctx := cmd.Context()
		resp, err := a.Chat(ctx, model, prompt)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to chat")
		}

		fmt.Println(resp)
	},
}
