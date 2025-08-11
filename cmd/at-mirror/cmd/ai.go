package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(aiCmd)
}

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI commands",
}
