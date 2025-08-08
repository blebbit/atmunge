package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "car",
	Short: "A tool to work with ATProto repo CAR files",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(mstCmd)
	rootCmd.AddCommand(unpackCmd)

	rootCmd.AddCommand(hackCmd)

	lsCmd.Aliases = []string{"list"}
}
