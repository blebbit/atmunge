package acct

import (
	"github.com/blebbit/atmunge/pkg/acct"
	"github.com/blebbit/atmunge/pkg/runtime"
	"github.com/spf13/cobra"
)

var phases []string

// acctSyncCmd represents the acctSync command
var acctSyncCmd = &cobra.Command{
	Use:   "sync [handle-or-did]",
	Short: "Sync an account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rt, err := runtime.NewRuntime(cmd.Context())
		if err != nil {
			return err
		}
		return acct.Sync(rt, args[0], phases)
	},
}

func init() {
	AcctCmd.AddCommand(acctSyncCmd)
	acctSyncCmd.Flags().StringSliceVar(&phases, "phases", []string{}, "phases to run (car, duckdb, blobs)")
}
