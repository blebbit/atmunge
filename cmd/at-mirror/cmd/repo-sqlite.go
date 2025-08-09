package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/blebbit/at-mirror/pkg/repo"
	"github.com/spf13/cobra"
)

var repoSqliteCmd = &cobra.Command{
	Use:   "sqlite [car file]",
	Short: "Convert a CAR file to an SQLite database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		carFile := args[0]
		output, _ := cmd.Flags().GetString("output")

		var dbPath string
		if output != "" {
			dbPath = output
		} else {
			dbPath = strings.TrimSuffix(carFile, filepath.Ext(carFile)) + ".db"
		}

		fmt.Printf("Converting %s to %s\n", carFile, dbPath)
		if err := repo.CarToSQLite(carFile, dbPath); err != nil {
			log.Fatalf("failed to convert CAR to SQLite: %v", err)
		}
	},
}

func init() {
	repoSqliteCmd.Flags().StringP("output", "o", "", "output file for sqlite db")
}
