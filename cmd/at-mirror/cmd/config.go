package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/blebbit/at-mirror/pkg/config"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show the current configuration",
	Long:  `Run the database migrations to update the schema.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return err
		}

		jsonBytes, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			fmt.Println("Error marshalling config to JSON:", err)
			return err
		}

		fmt.Println(string(jsonBytes))
		return nil
	},
}
