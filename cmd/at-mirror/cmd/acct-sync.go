package cmd

import (
	"github.com/blebbit/at-mirror/pkg/acct"

	"github.com/spf13/cobra"
)

// acctSyncCmd represents the acctSync command
var acctSyncCmd = &cobra.Command{
	Use:   "sync [handle-or-did]",
	Short: "Sync an account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return acct.Sync(args[0])
	},
}

func init() {
	acctCmd.AddCommand(acctSyncCmd)
}
