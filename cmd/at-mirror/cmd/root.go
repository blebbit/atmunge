package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "at-mirror",
	Short: "AT Mirror is a set of tools for sync'n and backfilling the AT Protocol network",
	Long: `AT Mirror is a set of tools for sync'n and backfilling the AT Protocol network
                made by Blebbit.
                Complete documentation is available at https://github.com/blebbit/at-mirror`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// Do Stuff Here
	// },
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
