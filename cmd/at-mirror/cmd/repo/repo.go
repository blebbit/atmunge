package repo

import (
	"github.com/spf13/cobra"
)

var RepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Commands for working with repositories",
	Long:  "Commands for working with repositories",
}

func init() {
	RepoCmd.AddCommand(repoHackCmd)
	RepoCmd.AddCommand(repoInspectCmd)
	RepoCmd.AddCommand(repoLsCmd)
	RepoCmd.AddCommand(repoMstCmd)
	RepoCmd.AddCommand(repoSyncCmd)
	RepoCmd.AddCommand(repoUnpackCmd)
	RepoCmd.AddCommand(repoSqliteCmd)
	RepoCmd.AddCommand(repoDuckDBCmd)
}
