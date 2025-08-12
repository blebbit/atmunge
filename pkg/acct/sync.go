package acct

import "fmt"

func Sync(handleOrDID string, phase string) error {
	fmt.Printf("Syncing account: %s from phase %s\n", handleOrDID, phase)
	return nil
}
