package acct

import "fmt"

func Sync(handleOrDID string) error {
	fmt.Printf("Syncing account: %s\n", handleOrDID)
	return nil
}
