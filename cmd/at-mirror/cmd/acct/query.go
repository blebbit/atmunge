package acct

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blebbit/at-mirror/pkg/acct"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var querySQLNames []string
var queryOutputDest string

func init() {
	AcctCmd.AddCommand(acctQueryCmd)
	acctQueryCmd.Flags().StringSliceVarP(&querySQLNames, "sql", "s", []string{}, "name of the adhoc query to run, or a raw SQL query")
	acctQueryCmd.Flags().StringVarP(&queryOutputDest, "output", "o", "-", "output file for results as JSON, or - for stdout")
	acctQueryCmd.RegisterFlagCompletionFunc("sql", getValidArgs("acct/query"))
}

var acctQueryCmd = &cobra.Command{
	Use:   "query <handle-or-did|list|view>",
	Short: "Run adhoc queries on an account's data in DuckDB",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "list":
			listValidSQLs("acct/query")
			return
		case "view":
			if len(args) < 2 {
				log.Fatal().Msg("view requires a second argument")
			}
			viewSQL("acct/query", args[1])
			return
		}

		ctx := cmd.Context()
		rt, err := runtime.NewRuntime(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create runtime")
		}

		handleOrDID := args[0]
		did, _, err := rt.ResolveDid(ctx, handleOrDID)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to resolve %s", handleOrDID)
		}

		dbPath := filepath.Join(rt.Cfg.RepoDataDir, did, "repo.duckdb")

		results, err := acct.Query(ctx, dbPath, querySQLNames)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to query account")
		}

		if queryOutputDest != "" && results != nil {
			jsonOutput, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				log.Fatal().Err(err).Msg("failed to marshal results to JSON")
			}

			if queryOutputDest == "-" {
				fmt.Println(string(jsonOutput))
			} else {
				if err := os.WriteFile(queryOutputDest, jsonOutput, 0644); err != nil {
					log.Fatal().Err(err).Msgf("failed to write output to %s", queryOutputDest)
				}
				log.Info().Str("path", queryOutputDest).Msg("wrote results to file")
			}
		}
	},
}
