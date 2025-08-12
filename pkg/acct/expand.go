package acct

import "fmt"

func Expand(handleOrDID string, levels int) error {
	fmt.Printf("Expanding account: %s, levels: %d\n", handleOrDID, levels)
	return nil
}
