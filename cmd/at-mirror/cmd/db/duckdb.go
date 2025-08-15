package db

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/spf13/cobra"
)

// duckDBCmd represents the duckdb command
var duckDBCmd = &cobra.Command{
	Use:   "duckdb <handle-or-did>",
	Short: "Open a DuckDB shell to an account's database",
	Long:  `Open a DuckDB shell to an account's database.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		rt, err := runtime.NewRuntime(cmd.Context())
		if err != nil {
			fmt.Printf("failed to create runtime: %s\n", err)
			return
		}

		did, _, err := rt.ResolveDid(cmd.Context(), args[0])
		if err != nil {
			fmt.Printf("failed to resolve did: %s\n", err)
			return
		}

		dbPath := filepath.Join(rt.Cfg.RepoDataDir, did, "repo.duckdb")

		// Check if the database file exists
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			fmt.Printf("database for %s not found at %s\n", args[0], dbPath)
			return
		}

		fmt.Printf("opening %s\n", dbPath)
		c := exec.Command("duckdb", dbPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			fmt.Printf("error: %s\n", err)
		}
	},
}

func init() {
	DBCmd.AddCommand(duckDBCmd)
}
