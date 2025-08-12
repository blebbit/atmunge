package cmd

import (
	"github.com/spf13/cobra"
)

// acctCmd represents the acct command
var acctCmd = &cobra.Command{
	Use:   "acct",
	Short: "Commands for working with accounts",
}

func init() {
	rootCmd.AddCommand(acctCmd)
}
