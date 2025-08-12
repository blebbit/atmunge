package acct

import (
	"github.com/blebbit/at-mirror/pkg/acct"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/spf13/cobra"
)

var phase string

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
		return acct.Sync(rt, args[0], phase)
	},
}

func init() {
	AcctCmd.AddCommand(acctSyncCmd)
	acctSyncCmd.Flags().StringVar(&phase, "phase", "", "phase to continue from (car, duckdb, blobs)")
}
