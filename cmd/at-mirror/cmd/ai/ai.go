package ai

import (
	"github.com/spf13/cobra"
)

func init() {
	AICmd.PersistentFlags().String("model", "gemma3:4b", "The model to use")
	AICmd.PersistentFlags().String("prompt", "", "The system prompt to use")
}

var AICmd = &cobra.Command{
	Use:   "ai",
	Short: "AI commands",
}
