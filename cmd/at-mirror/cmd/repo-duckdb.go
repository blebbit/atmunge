package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/blebbit/at-mirror/pkg/repo"
	"github.com/spf13/cobra"
)

var repoDuckDBCmd = &cobra.Command{
	Use:   "duckdb [car file]",
	Short: "Convert a CAR file to a DuckDB database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		carFile := args[0]
		output, _ := cmd.Flags().GetString("output")

		var dbPath string
		if output != "" {
			dbPath = output
		} else {
			dbPath = strings.TrimSuffix(carFile, filepath.Ext(carFile)) + ".duckdb"
		}

		fmt.Printf("Converting %s to %s\n", carFile, dbPath)
		if err := repo.CarToDuckDB(carFile, dbPath); err != nil {
			log.Fatalf("failed to convert CAR to DuckDB: %v", err)
		}
	},
}

func init() {
	repoDuckDBCmd.Flags().StringP("output", "o", "", "output file for duckdb")
}
