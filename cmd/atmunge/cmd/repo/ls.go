package repo

import (
	"fmt"
	"log"
	"os"

	"github.com/blebbit/atmunge/pkg/repo"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
)

var repoLsCmd = &cobra.Command{
	Use:     "ls [car file]",
	Short:   "List records in a CAR file",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"list"},
	Run: func(cmd *cobra.Command, args []string) {
		carFile := args[0]
		f, err := os.Open(carFile)
		if err != nil {
			log.Fatalf("failed to open car file: %v", err)
		}
		defer f.Close()

		r, err := repo.ReadRepoFromCar(f)
		if err != nil {
			log.Fatalf("failed to read repo from car: %v", err)
		}

		err = r.MST.Walk(func(k []byte, v cid.Cid) error {
			fmt.Printf("%s\t%s\n", string(k), v.String())
			return nil
		})
		if err != nil {
			log.Fatalf("failed to walk repo: %v", err)
		}
	},
}
