package repo

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/blebbit/at-mirror/pkg/repo"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/spf13/cobra"
)

var repoSyncCmd = &cobra.Command{
	Use:   "sync [account]",
	Short: "Sync a repo from a PDS",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		account := args[0]

		rt, err := runtime.NewRuntime(cmd.Context())
		if err != nil {
			log.Fatalf("failed to create runtime: %v", err)
		}

		plcInfo, err := repo.GetPlcInfo(account)
		if err != nil {
			log.Fatalf("failed to get PLC info: %v", err)
		}

		targetDID := plcInfo.DID
		pdsHost := plcInfo.PDSHost
		fmt.Printf("Syncing for %s (DID: %s) from PDS: %s\n", plcInfo.Handle, targetDID, pdsHost)

		dataDir := rt.Cfg.RepoDataDir
		if err := os.MkdirAll(dataDir, 0o755); err != nil {
			log.Fatalf("failed to create data directory: %v", err)
		}
		localCarFile := fmt.Sprintf("%s/%s.car", dataDir, targetDID)

		blockstoreMem, sinceTID, err := repo.LoadLocalCar(localCarFile)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Fatalf("failed to load local CAR for %s: %v", targetDID, err)
			}
			// if the file doesn't exist, we can continue, it's a new repo
		}
		if sinceTID == "" {
			fmt.Println("## No local repo found or no rev found. Performing initial sync...")
		}

		updateCarData, err := repo.GetRepo(pdsHost, targetDID, sinceTID)
		if err != nil {
			log.Fatalf("failed to fetch repo data: %v", err)
		}
		if len(updateCarData) == 0 {
			fmt.Println("No updates (empty response).")
			return
		}
		fmt.Printf("Fetched %d bytes of CAR data. Merging...\n", len(updateCarData))

		newRootCid, newestRev, _, err := repo.MergeUpdate(blockstoreMem, updateCarData)
		if err != nil {
			log.Fatalf("failed to merge update: %v", err)
		}

		if newestRev == "" {
			fmt.Println("No new commits in fetched data. Nothing to do.")
			return
		}

		if err := repo.WriteCar(localCarFile, newRootCid, blockstoreMem); err != nil {
			log.Fatalf("failed to write CAR file: %v", err)
		}

		fmt.Printf("Successfully synced %s. New root CID: %s, newest rev: %s\n", targetDID, newRootCid, newestRev)
		fmt.Printf("Local CAR file written to: %s\n", localCarFile)

		fmt.Println("Syncing blobs...")
		repo.SyncBlobs(pdsHost, targetDID, dataDir)
	},
}
