package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bluesky-social/indigo/atproto/data"
	"github.com/bluesky-social/indigo/atproto/repo"
	"github.com/bluesky-social/indigo/atproto/repo/mst"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car/v2"
	"github.com/spf13/cobra"
)

var hackCmd = &cobra.Command{
	Use:   "hack [args...]",
	Short: "developer hack command, change implementation as needed",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		carFile := args[0]
		f, err := os.Open(carFile)
		if err != nil {
			log.Fatalf("failed to open car file: %w", err)
		}
		defer f.Close()

		cr, err := car.NewBlockReader(f)
		if err != nil {
			log.Fatalf("failed to create block reader: %w", err)
		}

		fmt.Println("BR: ", cr.Version, cr.Roots)

		blocks := make(map[string][]byte)

		bs := repo.NewTinyBlockstore()
		for {
			blk, err := cr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatalf("failed to read block: %w", err)
			}

			blocks[blk.Cid().String()] = blk.RawData()
			if err := bs.Put(ctx, blk); err != nil {
				log.Fatalf("failed to put block in blockstore: %v", err)
			}
		}

		fmt.Println("Blocks:", len(blocks))

		commitBlock, err := bs.Get(ctx, cr.Roots[0])
		if err != nil {
			log.Fatalf("commit block not found in CAR: %v", err)
		}

		var commit repo.Commit
		if err := commit.UnmarshalCBOR(bytes.NewReader(commitBlock.RawData())); err != nil {
			log.Fatalf("parsing commit block from CAR file: %v", err)
		}
		if err := commit.VerifyStructure(); err != nil {
			log.Fatalf("parsing commit block from CAR file: %v", err)
		}

		tree, err := mst.LoadTreeFromStore(ctx, bs, commit.Data)
		if err != nil {
			log.Fatalf("reading MST from CAR file: %v", err)
		}
		clk := syntax.ClockFromTID(syntax.TID(commit.Rev))
		r := &repo.Repo{
			DID:         syntax.DID(commit.DID), // NOTE: VerifyStructure() already checked DID syntax
			Clock:       &clk,
			MST:         *tree,
			RecordStore: bs, // TODO: put just records in a smaller blockstore?
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
