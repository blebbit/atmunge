package backfill

import "github.com/spf13/cobra"

var (
	BackfillParallel int
	BackfillStart    string
)

var BackfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "Commands for backfilling from data sources",
	Long:  "Commands for backfilling from data sources",
}
