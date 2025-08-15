package acct

import (
	"github.com/blebbit/at-mirror/pkg/acct"

	"github.com/spf13/cobra"
)

// acctAnalyzeCmd represents the acctAnalyze command
var acctAnalyzeCmd = &cobra.Command{
	Use:   "analyze [handle-or-did]",
	Short: "Analyze an account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return acct.Analyze(args[0])
	},
}

func init() {
	AcctCmd.AddCommand(acctAnalyzeCmd)
}
