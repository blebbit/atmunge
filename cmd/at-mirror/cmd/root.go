package cmd

import (
	"fmt"
	"os"

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
	rootCmd.AddCommand(aiCmd)
	rootCmd.AddCommand(backfillCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(plcCmd)
	rootCmd.AddCommand(repoCmd)
	rootCmd.AddCommand(runCmd)
}

var (
	backfillParallel int
	backfillStart    string
)

var backfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "Commands for backfilling from data sources",
	Long:  "Commands for backfilling from data sources",
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Commands for working with the database",
	Long:  "Commands for working with the database",
}

var plcCmd = &cobra.Command{
	Use:   "plc",
	Short: "Commands for working with the PLC features",
	Long:  "Commands for working with the PLC features",
}
