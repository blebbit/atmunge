package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

type CarInfo struct {
	Roots   []cid.Cid
	Commits []Commit
}

type Commit struct {
	CID     cid.Cid
	Version int64    `cbor:"version"`
	Data    cid.Cid  `cbor:"data"`
	Rev     string   `cbor:"rev"`
	Prev    *cid.Cid `cbor:"prev"`
	DID     string   `cbor:"did"`
	Sig     []byte   `cbor:"sig"`
}

// GetCarInfo reads a CAR file and returns information about its roots and commits.
func GetCarInfo(filePath string) (*CarInfo, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open car file: %w", err)
	}
	defer f.Close()

	br, err := car.NewBlockReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create block reader: %w", err)
	}

	info := &CarInfo{
		Roots: br.Roots,
	}

	if len(br.Roots) == 0 {
		return info, nil
	}
	commitCid := br.Roots[0]

	// Reset file reader to search for the commit block from the beginning
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek in car file: %w", err)
	}

	// Re-initialize block reader
	br, err = car.NewBlockReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to re-create block reader: %w", err)
	}

	for {
		blk, err := br.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed reading block from CAR: %w", err)
		}

		if blk.Cid().Equals(commitCid) {
			var commit Commit
			if err := cbor.Unmarshal(blk.RawData(), &commit); err == nil {
				commit.CID = blk.Cid()
				info.Commits = append(info.Commits, commit)
			}
			// Found the root commit, we can stop.
			// If there are multiple roots, this only gets the first one.
			break
		}
	}

	return info, nil
}

// ExtractRoot reads CAR data and returns the root CID as a string.
func ExtractRoot(carData []byte) (string, error) {
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

// GetPlcInfo fetches account information from the PLC directory.
func GetPlcInfo(account string) (*PlcInfo, error) {
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

// LoadLocalCar loads an existing CAR file, returning its blocks and the latest commit TID.
func LoadLocalCar(filePath string) (map[cid.Cid][]byte, string, error) {
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

// MergeUpdate reads blocks from an update CAR, adds them to the block map,
// and returns the new root CID and the latest commit TID from the update.
func MergeUpdate(blockstoreMem map[cid.Cid][]byte, updateCarData []byte) (cid.Cid, string, error) {
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

// WriteCar writes the given blocks to a CAR file, atomically replacing the destination file.
func WriteCar(filePath string, rootCid cid.Cid, blockstoreMem map[cid.Cid][]byte) error {
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
