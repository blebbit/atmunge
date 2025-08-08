package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/blebbit/at-mirror/pkg/repo"
	"github.com/bluesky-social/indigo/atproto/data"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
)

var unpackCmd = &cobra.Command{
	Use:   "unpack [car file]",
	Short: "Unpack records from a CAR file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		carFile := args[0]
		outputDir, _ := cmd.Flags().GetString("output")

		f, err := os.Open(carFile)
		if err != nil {
			log.Fatalf("failed to open car file: %w", err)
		}
		defer f.Close()

		r, err := repo.ReadRepoFromCar(f)
		if err != nil {
			log.Fatalf("failed to read repo from car: %w", err)
		}

		if outputDir == "" {
			outputDir = string(r.DID)
		}

		ctx := context.Background()
		err = r.MST.Walk(func(k []byte, v cid.Cid) error {
			col, rkey, err := syntax.ParseRepoPath(string(k))
			if err != nil {
				return err
			}
			recBytes, _, err := r.GetRecordBytes(ctx, col, rkey)
			if err != nil {
				return err
			}

			rec, err := data.UnmarshalCBOR(recBytes)
			if err != nil {
				return err
			}

			recPath := filepath.Join(outputDir, string(k)+".json")
			fmt.Printf("%s\n", recPath)
			err = os.MkdirAll(filepath.Dir(recPath), os.ModePerm)
			if err != nil {
				return err
			}
			recJSON, err := json.MarshalIndent(rec, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(recPath, recJSON, 0666); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			log.Fatalf("failed to walk repo: %w", err)
		}
	},
}

func init() {
	unpackCmd.Flags().StringP("output", "o", "", "output directory")
}
