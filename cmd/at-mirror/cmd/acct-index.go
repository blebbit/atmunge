package cmd

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

var indexNames []string
var adhocNames []string
var outputDest string

func init() {
	acctCmd.AddCommand(acctIndexCmd)
	acctIndexCmd.Flags().StringSliceVar(&indexNames, "index", []string{}, "name of the index to run")
	acctIndexCmd.Flags().StringSliceVar(&adhocNames, "adhoc", []string{}, "name of the adhoc query to run")
	acctIndexCmd.Flags().StringVarP(&outputDest, "output", "o", "", "output file for results as JSON, or - for stdout")
}

var acctIndexCmd = &cobra.Command{
	Use:   "index <handle-or-did>",
	Short: "Index an account's data in DuckDB",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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

		dbPath := filepath.Join(rt.Cfg.RepoDataDir, did+".duckdb")

		indexer := acct.NewIndexer()
		results, err := indexer.Index(ctx, dbPath, indexNames, adhocNames)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to index account")
		}

		if outputDest != "" && results != nil {
			jsonOutput, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				log.Fatal().Err(err).Msg("failed to marshal results to JSON")
			}

			if outputDest == "-" {
				fmt.Println(string(jsonOutput))
			} else {
				if err := os.WriteFile(outputDest, jsonOutput, 0644); err != nil {
					log.Fatal().Err(err).Msgf("failed to write output to %s", outputDest)
				}
				log.Info().Str("path", outputDest).Msg("wrote results to file")
			}
		} else if outputDest == "" {
			fmt.Println("Indexing complete")
		}
	},
}
