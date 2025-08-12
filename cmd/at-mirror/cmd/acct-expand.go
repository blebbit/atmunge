package cmd

import (
	"github.com/blebbit/at-mirror/pkg/acct"

	"github.com/spf13/cobra"
)

// acctExpandCmd represents the acctExpand command
var acctExpandCmd = &cobra.Command{
	Use:   "expand [handle-or-did]",
	Short: "Expand an account's social graph",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		levels, _ := cmd.Flags().GetInt("levels")
		return acct.Expand(args[0], levels)
	},
}

func init() {
	acctCmd.AddCommand(acctExpandCmd)
	acctExpandCmd.Flags().IntP("levels", "l", 1, "Number of levels to expand")
}
