package acct

import (
	"fmt"
	"path/filepath"

	"github.com/blebbit/atmunge/pkg/acct"
	"github.com/blebbit/atmunge/pkg/runtime"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var indexNames []string

func init() {
	AcctCmd.AddCommand(acctIndexCmd)
	acctIndexCmd.Flags().StringSliceVar(&indexNames, "index", []string{}, "name of the index to run")
	acctIndexCmd.RegisterFlagCompletionFunc("index", getValidArgs("acct/index"))
}

var acctIndexCmd = &cobra.Command{
	Use:   "index <handle-or-did|list|view>",
	Short: "Index an account's data in DuckDB",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "list":
			listValidSQLs("acct/index")
			return
		case "view":
			if len(args) < 2 {
				log.Fatal().Msg("view requires a second argument")
			}
			viewSQL("acct/index", args[1])
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

		err = acct.Index(ctx, dbPath, indexNames)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to index account")
		}

		fmt.Println("Indexing complete")
	},
}
