package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mstCmd = &cobra.Command{
	Use:   "mst [car file]",
	Short: "Show repo MST structure",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mst not implemented yet")
	},
}
