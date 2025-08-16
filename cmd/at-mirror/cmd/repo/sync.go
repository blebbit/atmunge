package repo

import (
	"log"

	"github.com/blebbit/at-mirror/pkg/acct"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/spf13/cobra"
)

var repoSyncCmd = &cobra.Command{
	Use:   "sync [account]",
	Short: "Sync a repo from a PDS",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		account := args[0]

		rt, err := runtime.NewRuntime(cmd.Context())
		if err != nil {
			log.Fatalf("failed to create runtime: %v", err)
		}

		// The `acct.Sync` function handles car, duckdb, and blob syncing.
		// We can specify which phases to run, by default it runs all.
		if err := acct.Sync(rt, account, nil); err != nil {
			log.Fatalf("failed to sync account %s: %v", account, err)
		}
	},
}
