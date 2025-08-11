package ai

import (
	"context"
	"fmt"
)

func (a *AI) Topics(ctx context.Context, uri string) error {
	fmt.Println("getting topics for post: ", uri)
	return nil
}
