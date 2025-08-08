package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var unpackCmd = &cobra.Command{
	Use:   "unpack [car file]",
	Short: "Extract records from a CAR file as a directory of JSON files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("unpack not implemented yet")
	},
}
