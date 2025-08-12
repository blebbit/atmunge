package cmd

import (
	"fmt"
	"os"

	"github.com/blebbit/at-mirror/cmd/at-mirror/cmd/acct"
	"github.com/blebbit/at-mirror/cmd/at-mirror/cmd/ai"
	"github.com/blebbit/at-mirror/cmd/at-mirror/cmd/backfill"
	"github.com/blebbit/at-mirror/cmd/at-mirror/cmd/db"
	"github.com/blebbit/at-mirror/cmd/at-mirror/cmd/plc"
	"github.com/blebbit/at-mirror/cmd/at-mirror/cmd/repo"
	"github.com/spf13/cobra"
)

const rootLong = `
AT Mirror is a set of tools for backfilling and mirroring the AT Protocol network

Documentation and source code: https://github.com/blebbit/at-mirror
`

var rootCmd = &cobra.Command{
	Use:   "at-mirror",
	Short: "AT Mirror is a set of tools for backfilling and mirroring the AT Protocol network",
	Long:  rootLong,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(acct.AcctCmd)
	rootCmd.AddCommand(ai.AICmd)
	rootCmd.AddCommand(backfill.BackfillCmd)
	rootCmd.AddCommand(db.DBCmd)
	rootCmd.AddCommand(plc.PLCCmd)
	rootCmd.AddCommand(repo.RepoCmd)
	rootCmd.AddCommand(runCmd)
}
