package acct

import (
	"fmt"
	"path/filepath"

	"github.com/blebbit/at-mirror/pkg/acct"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var indexNames []string

func init() {
	AcctCmd.AddCommand(acctIndexCmd)
	acctIndexCmd.Flags().StringSliceVar(&indexNames, "index", []string{}, "name of the index to run")
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

		err = acct.Index(ctx, dbPath, indexNames)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to index account")
		}

		fmt.Println("Indexing complete")
	},
}
