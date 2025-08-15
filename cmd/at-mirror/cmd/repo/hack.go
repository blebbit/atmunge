package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/bluesky-social/indigo/atproto/data"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
)

var repoHackCmd = &cobra.Command{
	Use:   "hack [args...]",
	Short: "developer hack command, change implementation as needed",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		carFile := args[0]
		r, err := loadCar(ctx, carFile)
		if err != nil {
			log.Fatalf("failed to load CAR file: %v", err)
		}

		commit, err := r.Commit()
		if err != nil {
			log.Fatalf("failed to get commit: %v", err)
		}

		fmt.Println(r.DID, syntax.TID(commit.Rev).Time())

		fmt.Println("Repository loaded successfully")

		err = r.MST.Walk(func(k []byte, v cid.Cid) error {
			s := strings.Split(string(k), "/")
			fmt.Printf("%v\t%s\n", s, v.String())
			b, e := r.RecordStore.Get(ctx, v)
			if e != nil {
				log.Fatalf("failed to get block from record store: %v", e)
			}

			rec, err := data.UnmarshalCBOR(b.RawData())
			if err != nil {
				return err
			}

			// recPath := topDir + "/" + string(k)
			// fmt.Printf("%s.json\n", recPath)
			// err = os.MkdirAll(filepath.Dir(recPath), os.ModePerm)
			// if err != nil {
			// 	return err
			// }
			recJSON, err := json.MarshalIndent(rec, "", "  ")
			if err != nil {
				return err
			}

			fmt.Printf("%s\n\n", string(recJSON))
			return nil
		})
		if err != nil {
			log.Fatalf("failed to read records from repo CAR file: %v", err)
		}
	},
}
