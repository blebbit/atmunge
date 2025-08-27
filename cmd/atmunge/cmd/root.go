package cmd

import (
	"fmt"
	"os"

	"github.com/blebbit/atmunge/cmd/atmunge/cmd/acct"
	"github.com/blebbit/atmunge/cmd/atmunge/cmd/ai"
	"github.com/blebbit/atmunge/cmd/atmunge/cmd/backfill"
	"github.com/blebbit/atmunge/cmd/atmunge/cmd/db"
	"github.com/blebbit/atmunge/cmd/atmunge/cmd/plc"
	"github.com/blebbit/atmunge/cmd/atmunge/cmd/repo"
	"github.com/spf13/cobra"
)

const rootLong = `
AT Mirror is a set of tools for backfilling and mirroring the AT Protocol network

Documentation and source code: https://github.com/blebbit/atmunge
`

var rootCmd = &cobra.Command{
	Use:   "atmunge",
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
