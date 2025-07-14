package cmd

import (
	"github.com/spf13/cobra"
)

var (
	repoBackfillParallel    int
	repoBackfillStart       int
	repoBackfillEnd         int
	repoBackfillRetryErrors bool
)

func init() {
	repoCmd.AddCommand(repoBackfillCmd)
	repoBackfillCmd.PersistentFlags().IntVarP(&repoBackfillParallel, "parallel", "p", 42, "Number of parallel workers to use for backfilling")
	repoBackfillCmd.PersistentFlags().IntVarP(&repoBackfillStart, "start", "s", 0, "Start backfilling from this randomized ID list position")
	repoBackfillCmd.PersistentFlags().IntVarP(&repoBackfillEnd, "end", "e", -1, "End backfilling at this randomized ID list position (-1 means no limit)")
	repoBackfillCmd.PersistentFlags().BoolVarP(&repoBackfillRetryErrors, "retry-errors", "r", false, "Retry entries that had unhandled errors during backfill")
}

var repoBackfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "Commands for backfilling repos",
	Long:  "Commands for backfilling repos",
}
