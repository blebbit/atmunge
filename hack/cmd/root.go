package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hack",
	Short: "A collection of hack scripts",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(hackCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(mstCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(unpackCmd)
}
