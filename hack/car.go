package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	cbor "github.com/fxamacker/cbor/v2"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	car "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/blockstore"
)

// GetRepo fetches a repo's CAR file from a PDS.
// NOTE: We now pass the most recent commit TID (rev) as the `since` parameter
// instead of the root CID.
func GetRepo(pdsHost, did, since string) ([]byte, error) {
	endpoint, _ := url.Parse(pdsHost)
	endpoint.Path = "/xrpc/com.atproto.sync.getRepo"
	queryParams := url.Values{}
	queryParams.Set("did", did)
	if since != "" {
		queryParams.Set("since", since)
	}
	endpoint.RawQuery = queryParams.Encode()

	fmt.Printf("Fetching from: %s\n", endpoint.String())
	resp, err := http.Get(endpoint.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad response status: %s, body: %s", resp.Status, string(body))
	}
	return io.ReadAll(resp.Body)
}

// extractRoot reads CAR data and returns the root CID as a string.
func extractRoot(carData []byte) (string, error) {
	br, err := car.NewBlockReader(bytes.NewReader(carData))
	if err != nil {
		return "", err
	}
	roots := br.Roots
	if len(roots) == 0 {
		return "", fmt.Errorf("car file has no roots")
	}
	return roots[0].String(), nil
}

// tryExtractRev attempts to decode a block as CBOR and return a commit rev (TID) if present.
func tryExtractRev(raw []byte) (string, bool) {
	var m map[string]any
	if err := cbor.Unmarshal(raw, &m); err != nil {
		return "", false
	}
	// Heuristic: commit block contains keys: did, rev, data, sig (at least rev)
	if revVal, ok := m["rev"].(string); ok && revVal != "" {
		return revVal, true
	}
	return "", false
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <did>", os.Args[0])
	}
	targetDID := os.Args[1]

	pdsHost := "https://russula.us-west.host.bsky.network"
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0o755); err != nil { // ensure data dir exists
		log.Fatalf("Failed to create data directory: %v", err)
	}
	localCarFile := fmt.Sprintf("%s/%s.car", dataDir, targetDID)
	var sinceTID string // we will pass latest commit rev (TID) instead of root CID

	// Keep an in-memory map of blocks for (re)writing the full CAR.
	blockstoreMem := make(map[cid.Cid][]byte)

	// If a local CAR exists, load its rev (for diff) and all blocks.
	if f, err := os.Open(localCarFile); err == nil {
		fmt.Println("## Local repo found. Loading existing CAR...")
		br, err := car.NewBlockReader(f)
		if err != nil {
			f.Close()
			log.Fatalf("Failed to read existing CAR: %v", err)
		}
		var latestRev string
		for {
			blk, err := br.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				f.Close()
				log.Fatalf("Failed reading block from existing CAR: %v", err)
			}
			blockstoreMem[blk.Cid()] = blk.RawData()
			if rev, ok := tryExtractRev(blk.RawData()); ok {
				latestRev = rev
			}
		}
		f.Close()
		if latestRev != "" {
			sinceTID = latestRev
			fmt.Printf("Most recent commit TID (rev) in local CAR: %s\n", sinceTID)
		} else {
			fmt.Println("No commit rev found in existing CAR; full sync will be performed.")
		}
	} else if !os.IsNotExist(err) { // unexpected stat error
		log.Fatalf("Failed stating local CAR: %v", err)
	} else {
		fmt.Println("## No local repo found. Performing initial sync...")
	}

	// Fetch full or diff CAR from PDS.
	updateCarData, err := GetRepo(pdsHost, targetDID, sinceTID)
	if err != nil {
		log.Fatalf("Failed to fetch repo data: %v", err)
	}
	if len(updateCarData) == 0 {
		fmt.Println("No updates (empty response).")
		return
	}
	fmt.Printf("Fetched %d bytes of CAR data. Merging...\n", len(updateCarData))

	// Read the update CAR blocks and merge into in-memory map; record new root and most recent rev.
	updateBR, err := car.NewBlockReader(bytes.NewReader(updateCarData))
	if err != nil {
		log.Fatalf("Failed to parse fetched CAR: %v", err)
	}
	if len(updateBR.Roots) == 0 {
		log.Fatalf("Fetched CAR has no roots")
	}
	newRootCid := updateBR.Roots[0]
	var newestRev string
	for {
		blk, err := updateBR.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed reading block from fetched CAR: %v", err)
		}
		blockstoreMem[blk.Cid()] = blk.RawData()
		if rev, ok := tryExtractRev(blk.RawData()); ok {
			newestRev = rev
			fmt.Printf("Found commit block CID=%s rev=%s (treating as latest)\n", blk.Cid(), rev)
		}
	}
	if newestRev != "" {
		fmt.Printf("Most recent commit rev (TID) from fetched CAR: %s\n", newestRev)
	} else {
		fmt.Println("Warning: No commit rev found in fetched CAR; future incremental sync may not work as expected.")
	}

	// Write merged blocks into a fresh CAR file using the supported blockstore API.
	tempPath := localCarFile + ".tmp"
	_ = os.Remove(tempPath) // cleanup prior temp if any
	ctx := context.Background()

	// Open a new read-write CAR (always from scratch) at tempPath.
	rwbs, err := blockstore.OpenReadWrite(tempPath, []cid.Cid{newRootCid})
	if err != nil {
		log.Fatalf("Failed to open read-write CAR: %v", err)
	}

	// Convert map to slice of blocks and write.
	writeBuf := make([]blocks.Block, 0, len(blockstoreMem))
	for c, data := range blockstoreMem {
		b, err := blocks.NewBlockWithCid(data, c)
		if err != nil {
			log.Fatalf("Failed to construct block for %s: %v", c, err)
		}
		writeBuf = append(writeBuf, b)
	}
	if err := rwbs.PutMany(ctx, writeBuf); err != nil {
		log.Fatalf("Failed writing blocks: %v", err)
	}
	if err := rwbs.Finalize(); err != nil { // ensures complete CAR (header, index)
		log.Fatalf("Failed finalizing CAR: %v", err)
	}

	// Atomically replace old file.
	if err := os.Rename(tempPath, localCarFile); err != nil {
		log.Fatalf("Failed replacing CAR file: %v", err)
	}
	fmt.Println("Wrote merged CAR with root", newRootCid, "to", localCarFile)
}
