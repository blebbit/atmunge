package cmd

import (
	"fmt"

	"github.com/blebbit/at-mirror/pkg/ai"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	aiCmd.AddCommand(aiCompleteCmd)
	aiCompleteCmd.Flags().String("model", "llama3", "The model to use for generating completions")
}

var aiCompleteCmd = &cobra.Command{
	Use:   "complete <prompt>",
	Short: "Complete a string",
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
		resp, err := a.Complete(ctx, model, prompt)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to complete")
		}

		fmt.Println(resp)
	},
}
