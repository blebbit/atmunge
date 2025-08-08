package cmd

import (
	"fmt"
	"log"

	"github.com/blebbit/at-mirror/pkg/repo"
	"github.com/spf13/cobra"
)

var repoInspectCmd = &cobra.Command{
	Use:   "inspect [car file]",
	Short: "Show commit metadata from a CAR file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		carFile := args[0]
		info, err := repo.GetCarInfo(carFile)
		if err != nil {
			log.Fatalf("failed to get car info: %v", err)
		}

		fmt.Printf("Roots: %v\n", info.Roots)
		fmt.Println("Commits:")
		for _, commit := range info.Commits {
			fmt.Printf("  - CID: %s\n", commit.CID)
			fmt.Printf("    Rev: %s\n", commit.Rev)
			fmt.Printf("    Prev: %s\n", commit.Prev)
			fmt.Printf("    Data: %s\n", commit.Data)
			fmt.Printf("    DID: %s\n", commit.DID)
		}
	},
}
