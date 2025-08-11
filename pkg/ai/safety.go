package ai

import (
	"context"
	"fmt"
)

func (a *AI) Safety(ctx context.Context, uri string) error {
	fmt.Println("getting safety status for post: ", uri)
	return nil
}
