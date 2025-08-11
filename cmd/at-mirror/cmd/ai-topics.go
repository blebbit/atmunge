package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/blebbit/at-mirror/pkg/ai"
)

func init() {
	aiCmd.AddCommand(aiTopicsCmd)
}

var aiTopicsCmd = &cobra.Command{
	Use:   "topics <at-uri>",
	Short: "Get topics for a post",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}
		uri := args[0]
		log.Info().Msgf("getting topics for post: %s", uri)

		a, err := ai.NewAI()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create AI client")
		}

		ctx := cmd.Context()
		if err := a.Topics(ctx, uri); err != nil {
			log.Fatal().Err(err).Msg("failed to get topics")
		}

		fmt.Println("ok")
	},
}
