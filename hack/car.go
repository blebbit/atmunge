package main

import (
	"bytes"
	"context"
	"encoding/json"
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

type PlcInfo struct {
	DID      string `json:"did"`
	PDSHost  string `json:"pds"`
	Handle   string `json:"handle"`
	PlcTime  string `json:"plcTime"`
	LastTime string `json:"lastTime"`
}

// getPlcInfo fetches account information from the PLC directory.
func getPlcInfo(account string) (*PlcInfo, error) {
	plcURL := fmt.Sprintf("https://plc.blebbit.dev/info/%s", account)
	resp, err := http.Get(plcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info from PLC: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get account info from PLC: %s, body: %s", resp.Status, string(body))
	}

	var plcInfo PlcInfo
	if err := json.NewDecoder(resp.Body).Decode(&plcInfo); err != nil {
		return nil, fmt.Errorf("failed to decode PLC info response: %w", err)
	}
	return &plcInfo, nil
}

// loadLocalCar loads an existing CAR file, returning its blocks and the latest commit TID.
func loadLocalCar(filePath string) (map[cid.Cid][]byte, string, error) {
	blockstoreMem := make(map[cid.Cid][]byte)
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return blockstoreMem, "", nil // Not an error, just no local file
		}
		return nil, "", fmt.Errorf("failed stating local CAR: %w", err)
	}
	defer f.Close()

	fmt.Println("## Local repo found. Loading existing CAR...")
	br, err := car.NewBlockReader(f)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read existing CAR: %w", err)
	}

	var latestRev string
	for {
		blk, err := br.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed reading block from existing CAR: %w", err)
		}
		blockstoreMem[blk.Cid()] = blk.RawData()
		if rev, ok := tryExtractRev(blk.RawData()); ok {
			if latestRev == "" || rev > latestRev {
				latestRev = rev
			}
		}
	}

	if latestRev != "" {
		fmt.Printf("Most recent commit TID (rev) in local CAR: %s\n", latestRev)
	} else {
		fmt.Println("No commit rev found in existing CAR; full sync will be performed.")
	}

	return blockstoreMem, latestRev, nil
}

// mergeUpdate reads blocks from an update CAR, adds them to the block map,
// and returns the new root CID and the latest commit TID from the update.
func mergeUpdate(blockstoreMem map[cid.Cid][]byte, updateCarData []byte) (cid.Cid, string, error) {
	updateBR, err := car.NewBlockReader(bytes.NewReader(updateCarData))
	if err != nil {
		return cid.Undef, "", fmt.Errorf("failed to parse fetched CAR: %w", err)
	}
	if len(updateBR.Roots) == 0 {
		return cid.Undef, "", fmt.Errorf("fetched CAR has no roots")
	}
	newRootCid := updateBR.Roots[0]
	var newestRev string
	for {
		blk, err := updateBR.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cid.Undef, "", fmt.Errorf("failed reading block from fetched CAR: %w", err)
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
	return newRootCid, newestRev, nil
}

// writeCar writes the given blocks to a CAR file, atomically replacing the destination file.
func writeCar(filePath string, rootCid cid.Cid, blockstoreMem map[cid.Cid][]byte) error {
	tempPath := filePath + ".tmp"
	_ = os.Remove(tempPath) // cleanup prior temp if any
	ctx := context.Background()

	rwbs, err := blockstore.OpenReadWrite(tempPath, []cid.Cid{rootCid})
	if err != nil {
		return fmt.Errorf("failed to open read-write CAR: %w", err)
	}

	writeBuf := make([]blocks.Block, 0, len(blockstoreMem))
	for c, data := range blockstoreMem {
		b, err := blocks.NewBlockWithCid(data, c)
		if err != nil {
			return fmt.Errorf("failed to construct block for %s: %w", c, err)
		}
		writeBuf = append(writeBuf, b)
	}
	if err := rwbs.PutMany(ctx, writeBuf); err != nil {
		return fmt.Errorf("failed writing blocks: %w", err)
	}
	if err := rwbs.Finalize(); err != nil {
		return fmt.Errorf("failed finalizing CAR: %w", err)
	}

	if err := os.Rename(tempPath, filePath); err != nil {
		return fmt.Errorf("failed replacing CAR file: %w", err)
	}
	fmt.Println("Wrote merged CAR with root", rootCid, "to", filePath)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <account>", os.Args[0])
	}
	account := os.Args[1]

	plcInfo, err := getPlcInfo(account)
	if err != nil {
		log.Fatalf("failed to get PLC info: %v", err)
	}

	targetDID := plcInfo.DID
	pdsHost := plcInfo.PDSHost
	fmt.Printf("Syncing for %s (DID: %s) from PDS: %s\n", plcInfo.Handle, targetDID, pdsHost)

	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}
	localCarFile := fmt.Sprintf("%s/%s.car", dataDir, targetDID)

	blockstoreMem, sinceTID, err := loadLocalCar(localCarFile)
	if err != nil {
		log.Fatalf("failed to load local CAR: %v", err)
	}
	if sinceTID == "" {
		fmt.Println("## No local repo found or no rev found. Performing initial sync...")
	}

	updateCarData, err := GetRepo(pdsHost, targetDID, sinceTID)
	if err != nil {
		log.Fatalf("failed to fetch repo data: %v", err)
	}
	if len(updateCarData) == 0 {
		fmt.Println("No updates (empty response).")
		return
	}
	fmt.Printf("Fetched %d bytes of CAR data. Merging...\n", len(updateCarData))

	newRootCid, newestRev, err := mergeUpdate(blockstoreMem, updateCarData)
	if err != nil {
		log.Fatalf("failed to merge update: %v", err)
	}

	if newestRev == "" {
		fmt.Println("No new commits in fetched data. Nothing to do.")
		return
	}

	if err := writeCar(localCarFile, newRootCid, blockstoreMem); err != nil {
		log.Fatalf("failed to write CAR file: %v", err)
	}
}
