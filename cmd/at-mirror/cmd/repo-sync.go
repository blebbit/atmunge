package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/blebbit/at-mirror/pkg/repo"
	"github.com/spf13/cobra"
)

var repoSyncCmd = &cobra.Command{
	Use:   "sync [account]",
	Short: "Sync a repo from a PDS",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		account := args[0]

		plcInfo, err := repo.GetPlcInfo(account)
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

		blockstoreMem, sinceTID, err := repo.LoadLocalCar(localCarFile)
		if err != nil {
			log.Fatalf("failed to load local CAR: %v", err)
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

		newRootCid, newestRev, err := repo.MergeUpdate(blockstoreMem, updateCarData)
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
	},
}
