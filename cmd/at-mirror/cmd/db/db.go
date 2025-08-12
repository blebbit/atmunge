package db

import "github.com/spf13/cobra"

var DBCmd = &cobra.Command{
	Use:   "db",
	Short: "Commands for working with the database",
	Long:  "Commands for working with the database",
}
