package acct

import "fmt"

func Stats(handleOrDID string) error {
	fmt.Printf("Getting stats for account: %s\n", handleOrDID)
	return nil
}
