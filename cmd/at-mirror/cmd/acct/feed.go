package acct

import (
	"github.com/blebbit/atmunge/pkg/acct"

	"github.com/spf13/cobra"
)

// acctFeedCmd represents the acctFeed command
var acctFeedCmd = &cobra.Command{
	Use:   "feed [args...]",
	Short: "Get a feed for an account",
	RunE: func(cmd *cobra.Command, args []string) error {
		return acct.Feed(args)
	},
}

func init() {
	AcctCmd.AddCommand(acctFeedCmd)
}
