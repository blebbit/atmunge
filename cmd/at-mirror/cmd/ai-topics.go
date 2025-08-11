package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bleb-inc/at-mirror/pkg/ai"
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
		log.Infof("getting topics for post: %s", uri)

		a, err := ai.NewAI()
		if err != nil {
			log.Fatal(err)
		}

		ctx := cmd.Context()
		if err := a.Topics(ctx, uri); err != nil {
			log.Fatal(err)
		}

		fmt.Println("ok")
	},
}
