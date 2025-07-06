package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	// plcSnapshotCmd.PersistentFlags().StringP("filter", "f", "", "Filter the PLC snapshot for bad operations")
	rootCmd.AddCommand(dbCmd)
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Commands for working with the database",
	Long:  "Commands for working with the database",
}
