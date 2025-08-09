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

	indigoRepo "github.com/bluesky-social/indigo/atproto/repo"
	"github.com/bluesky-social/indigo/atproto/repo/mst"
	"github.com/bluesky-social/indigo/atproto/syntax"
)

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

type PlcInfo struct {
	DID      string `json:"did"`
	PDSHost  string `json:"pds"`
	Handle   string `json:"handle"`
	PlcTime  string `json:"plcTime"`
	LastTime string `json:"lastTime"`
}

// GetPlcInfo fetches account information from the PLC directory.
// this can probably come from the database, or even during lookup, we could pluck more than did
// for now, this means the repo commands work while the backfill command remains to be optimized
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

	// fmt.Printf("Fetching from: %s\n", endpoint.String())
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

func OpenCarBlockReader(filePath string) (*car.BlockReader, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open car file: %w", err)
	}
	defer f.Close()

	br, err := car.NewBlockReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create block reader: %w", err)
	}

	return br, nil
}

// GetCarInfo reads a CAR file and returns information about its roots and commits.
func GetCarInfo(filePath string) (*CarInfo, error) {
	br, err := OpenCarBlockReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CAR as block reader: %w", err)
	}

	info := &CarInfo{
		Roots: br.Roots,
	}

	if len(br.Roots) == 0 {
		return info, nil
	}
	commitCid := br.Roots[0]

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

// LoadLocalCar loads an existing CAR file, returning its blocks and the latest commit TID.
func LoadLocalCar(filePath string) (map[cid.Cid][]byte, string, error) {
	blockstoreMem := make(map[cid.Cid][]byte)

	br, err := OpenCarBlockReader(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open CAR as block reader: %w", err)
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

	return blockstoreMem, latestRev, nil
}

// MergeUpdate reads blocks from an update CAR, adds them to the block map,
// and returns the new root CID and the latest commit TID from the update.
func MergeUpdate(blockstoreMem map[cid.Cid][]byte, updateCarData []byte) (cid.Cid, string, map[cid.Cid][]byte, error) {
	newBlocks := make(map[cid.Cid][]byte)
	updateBR, err := car.NewBlockReader(bytes.NewReader(updateCarData))
	if err != nil {
		return cid.Undef, "", nil, fmt.Errorf("failed to parse fetched CAR: %w", err)
	}
	if len(updateBR.Roots) == 0 {
		// This can happen with an empty diff when a repo is up-to-date.
		// It's not an error, just means no new blocks.
		return cid.Undef, "", newBlocks, nil
	}
	newRootCid := updateBR.Roots[0]
	var newestRev string
	for {
		blk, err := updateBR.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cid.Undef, "", nil, fmt.Errorf("failed reading block from fetched CAR: %w", err)
		}
		blockstoreMem[blk.Cid()] = blk.RawData()
		newBlocks[blk.Cid()] = blk.RawData()
		if rev, ok := tryExtractRev(blk.RawData()); ok {
			newestRev = rev
		}
	}
	return newRootCid, newestRev, newBlocks, nil
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
	return nil
}

func ReadRepoFromCar(r io.Reader) (*indigoRepo.Repo, error) {
	ctx := context.Background()

	br, err := car.NewBlockReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create block reader: %w", err)
	}

	bs := indigoRepo.NewTinyBlockstore()
	for {
		blk, err := br.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read block: %w", err)
		}

		if err := bs.Put(ctx, blk); err != nil {
			return nil, fmt.Errorf("failed to put block in blockstore: %v", err)
		}
	}

	if len(br.Roots) == 0 {
		return nil, fmt.Errorf("car has no roots")
	}

	commitBlock, err := bs.Get(ctx, br.Roots[0])
	if err != nil {
		return nil, fmt.Errorf("commit block not found in CAR: %v", err)
	}

	var commit indigoRepo.Commit
	if err := commit.UnmarshalCBOR(bytes.NewReader(commitBlock.RawData())); err != nil {
		return nil, fmt.Errorf("parsing commit block from CAR file: %v", err)
	}
	if err := commit.VerifyStructure(); err != nil {
		return nil, fmt.Errorf("verifying commit block from CAR file: %v", err)
	}

	tree, err := mst.LoadTreeFromStore(ctx, bs, commit.Data)
	if err != nil {
		return nil, fmt.Errorf("reading MST from CAR file: %v", err)
	}
	clk := syntax.ClockFromTID(syntax.TID(commit.Rev))
	repo := &indigoRepo.Repo{
		DID:         syntax.DID(commit.DID), // NOTE: VerifyStructure() already checked DID syntax
		Clock:       &clk,
		MST:         *tree,
		RecordStore: bs,
	}

	return repo, nil
}
