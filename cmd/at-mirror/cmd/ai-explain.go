package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/blebbit/at-mirror/pkg/ai"
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
		uri := args[0]
		log.Info().Msgf("explaining post: %s", uri)

		a, err := ai.NewAI()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create AI client")
		}

		ctx := cmd.Context()
		if err := a.Explain(ctx, uri); err != nil {
			log.Fatal().Err(err).Msg("failed to explain")
		}

		fmt.Println("ok")
	},
}
