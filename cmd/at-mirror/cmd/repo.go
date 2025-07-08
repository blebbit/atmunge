package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	// plcSnapshotCmd.PersistentFlags().StringP("filter", "f", "", "Filter the PLC snapshot for bad operations")
	rootCmd.AddCommand(repoCmd)
}

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Commands for working with the repo features",
	Long:  "Commands for working with the repo features",
}
