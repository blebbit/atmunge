package cmd

import (
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Commands for working with repositories",
	Long:  "Commands for working with repositories",
}

func init() {
	repoCmd.AddCommand(repoHackCmd)
	repoCmd.AddCommand(repoInspectCmd)
	repoCmd.AddCommand(repoLsCmd)
	repoCmd.AddCommand(repoMstCmd)
	repoCmd.AddCommand(repoSyncCmd)
	repoCmd.AddCommand(repoUnpackCmd)
}
