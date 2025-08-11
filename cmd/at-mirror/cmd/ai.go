package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	aiCmd.PersistentFlags().String("model", "gemma3:4b", "The model to use")
	aiCmd.PersistentFlags().String("prompt", "", "The system prompt to use")
}

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI commands",
}
