package acct

import (
	"github.com/blebbit/at-mirror/pkg/acct"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// acctExpandCmd represents the acctExpand command
var acctExpandCmd = &cobra.Command{
	Use:   "expand [handle-or-did] [what-to-expand]",
	Short: "Expand an account's social graph",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		rt, err := runtime.NewRuntime(cmd.Context())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create runtime")
		}

		handleOrDID := args[0]
		what := args[1]

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
