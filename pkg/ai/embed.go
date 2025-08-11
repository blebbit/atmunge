package ai

import (
	"context"
	"fmt"
)

func (a *AI) Embed(ctx context.Context, uri string) error {
	fmt.Println("embedding post: ", uri)
	return nil
}
