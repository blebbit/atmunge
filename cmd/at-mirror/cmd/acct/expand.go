package acct

import (
	"fmt"

	"github.com/blebbit/at-mirror/pkg/acct"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// acctExpandCmd represents the acctExpand command
var acctExpandCmd = &cobra.Command{
	Use:   "expand [what-to-expand] [handle-or-did]",
	Short: "Expand an account's social graph",
	Long: `Expand an account's social graph.

You can list the available expansion targets with:
at-mirror acct expand list
`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && args[0] == "list" {
			fmt.Println("Available expansion targets:")
			for _, key := range acct.GetExpansionKeys() {
				fmt.Printf("- %s\n", key)
			}
			return
		}

		if len(args) != 2 {
			cmd.Help()
			return
		}

		rt, err := runtime.NewRuntime(cmd.Context())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create runtime")
		}

		what := args[0]
		handleOrDID := args[1]

		did, _, err := rt.ResolveDid(cmd.Context(), handleOrDID)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to resolve %s", handleOrDID)
		}

		err = acct.Expand(rt, did, what)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to expand account")
		}
	},
}

func init() {
	AcctCmd.AddCommand(acctExpandCmd)
}
