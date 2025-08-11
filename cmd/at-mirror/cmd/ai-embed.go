package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bleb-inc/at-mirror/pkg/ai"
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
		log.Infof("embedding post: %s", uri)

		a, err := ai.NewAI()
		if err != nil {
			log.Fatal(err)
		}

		ctx := cmd.Context()
		if err := a.Embed(ctx, uri); err != nil {
			log.Fatal(err)
		}

		fmt.Println("ok")
	},
}
