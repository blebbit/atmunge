package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bluesky-social/indigo/atproto/repo"
	"github.com/bluesky-social/indigo/atproto/repo/mst"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

type repoMSTOptions struct {
	fullCID bool
}

var mstCmd = &cobra.Command{
	Use:   "mst [car file]",
	Short: "Show repo MST structure",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		carFile := args[0]

		r, err := loadCar(ctx, carFile)
		if err != nil {
			log.Fatalf("failed to load CAR file: %v", err)
		}

		opts := repoMSTOptions{
			fullCID: false,
		}

		tree := treeprint.NewWithRoot(displayCID(r.MST.Root.CID, true, opts))

		if err := walkMST(ctx, r.MST.Root, r.RecordStore, tree, opts, 100); err != nil {
			log.Fatalf("failed to walk MST: %v", err)
		}

		// print tree
		fmt.Println(tree.String())
	},
}

func walkMST(ctx context.Context, node *mst.Node, bs repo.RepoBlockSource, tree treeprint.Tree, opts repoMSTOptions, depth int) error {

	if depth == 0 {
		return nil
	}

	// entries := tree.AddBranch("node.Entries")
	for _, entry := range node.Entries {
		if entry.Child != nil {
			exists, err := nodeExists(ctx, bs, entry.ChildCID)
			if err != nil {
				return err
			}
			subtree := tree.AddBranch(displayCID(entry.ChildCID, exists, opts))
			if exists {
				if err := walkMST(ctx, entry.Child, bs, subtree, opts, depth-1); err != nil {
					return err
				}
			}
		}

		if len(entry.Key) > 0 {
			tree.AddBranch(fmt.Sprintf("%s %s", entry.Key, displayCID(entry.Value, true, opts)))
		}
	}

	return nil
}

func displayEntryVal(entry *mst.EntryData, exists bool, opts repoMSTOptions) string {
	key := string(entry.KeySuffix)
	divider := " "
	if opts.fullCID {
		divider = "\n"
	}
	return strings.Repeat("∙", int(entry.PrefixLen)) + key + divider + displayCID(&entry.Value, exists, opts)
}

func displayCID(cid *cid.Cid, exists bool, opts repoMSTOptions) string {
	cidDisplay := cid.String()
	if !opts.fullCID {
		cidDisplay = "…" + string(cidDisplay[len(cidDisplay)-7:])
	}
	connector := "─◉"
	if !exists {
		connector = "─◌"
	}
	return "[" + cidDisplay + "]" + connector
}

func nodeExists(ctx context.Context, bs repo.RepoBlockSource, cid *cid.Cid) (bool, error) {
	if _, err := bs.Get(ctx, *cid); err != nil {
		if errors.Is(err, ipld.ErrNotFound{}) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
