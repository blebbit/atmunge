package acct

import (
	"github.com/blebbit/atmunge/pkg/acct"

	"github.com/spf13/cobra"
)

// acctStatsCmd represents the acctStats command
var acctStatsCmd = &cobra.Command{
	Use:   "stats [handle-or-did]",
	Short: "Get stats for an account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return acct.Stats(args[0])
	},
}

func init() {
	AcctCmd.AddCommand(acctStatsCmd)
}
