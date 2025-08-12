package repo

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"

	"github.com/bluesky-social/indigo/atproto/repo"
	"github.com/bluesky-social/indigo/atproto/repo/mst"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/ipld/go-car/v2"
)

func loadCar(ctx context.Context, carFile string) (*repo.Repo, error) {
	f, err := os.Open(carFile)
	if err != nil {
		log.Fatalf("failed to open car file: %w", err)
	}
	defer f.Close()

	cr, err := car.NewBlockReader(f)
	if err != nil {
		log.Fatalf("failed to create block reader: %w", err)
	}

	bs := repo.NewTinyBlockstore()
	for {
		blk, err := cr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("failed to read block: %w", err)
		}

		if err := bs.Put(ctx, blk); err != nil {
			log.Fatalf("failed to put block in blockstore: %v", err)
		}
	}

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

	return r, nil
}
