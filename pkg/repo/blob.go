package repo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// ListBlobs fetches the list of all blob CIDs for a given repo, handling pagination.
func ListBlobs(pdsHost, did string) ([]string, error) {
	var allCids []string
	var cursor string

	for {
		endpoint, _ := url.Parse(pdsHost)
		endpoint.Path = "/xrpc/com.atproto.sync.listBlobs"
		queryParams := url.Values{}
		queryParams.Set("did", did)
		queryParams.Set("limit", "1000") // Adjust limit as needed
		if cursor != "" {
			queryParams.Set("cursor", cursor)
		}
		endpoint.RawQuery = queryParams.Encode()

		resp, err := http.Get(endpoint.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("bad response status: %s, body: %s", resp.Status, string(body))
		}

		var listResp struct {
			Cursor *string  `json:"cursor"`
			Cids   []string `json:"cids"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
			return nil, fmt.Errorf("failed to decode listBlobs response: %w", err)
		}

		allCids = append(allCids, listResp.Cids...)

		if listResp.Cursor == nil || *listResp.Cursor == "" {
			break
		}
		cursor = *listResp.Cursor
	}

	return allCids, nil
}

// GetBlob fetches a single blob.
func GetBlob(pdsHost, did, cid string) ([]byte, error) {
	endpoint, _ := url.Parse(pdsHost)
	endpoint.Path = "/xrpc/com.atproto.sync.getBlob"
	queryParams := url.Values{}
	queryParams.Set("did", did)
	queryParams.Set("cid", cid)
	endpoint.RawQuery = queryParams.Encode()

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

// SyncBlobs fetches and saves blobs for a repo.
func SyncBlobs(pdsHost, did, dataDir string) {
	cids, err := ListBlobs(pdsHost, did)
	if err != nil {
		log.Printf("failed to list blobs: %v", err)
		return
	}

	if len(cids) == 0 {
		fmt.Println("No blobs found for repo.")
		return
	}

	fmt.Printf("Found %d blobs for repo. Checking for missing blobs...\n", len(cids))

	blobsDir := filepath.Join(dataDir, did, "blobs")
	if err := os.MkdirAll(blobsDir, 0o755); err != nil {
		log.Printf("failed to create blobs directory: %v", err)
		return
	}

	for _, c := range cids {
		blobPath := filepath.Join(blobsDir, c+".blob")
		if _, err := os.Stat(blobPath); err == nil {
			// file exists, skip
			continue
		}

		blobData, err := GetBlob(pdsHost, did, c)
		if err != nil {
			log.Printf("failed to get blob %s: %v", c, err)
			continue // continue to next blob
		}

		if err := os.WriteFile(blobPath, blobData, 0o644); err != nil {
			log.Printf("failed to write blob %s: %v", c, err)
			continue // continue to next blob
		}
		fmt.Printf("Saved blob: %s\n", blobPath)
	}
}
