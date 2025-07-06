package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	// plcSnapshotCmd.PersistentFlags().StringP("filter", "f", "", "Filter the PLC snapshot for bad operations")
	rootCmd.AddCommand(plcCmd)
}

var plcCmd = &cobra.Command{
	Use:   "plc",
	Short: "Commands for working with the PLC features",
	Long:  "Commands for working with the PLC features",
}
