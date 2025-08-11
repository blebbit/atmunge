package ai

import (
	"context"
	"fmt"
)

func (a *AI) Explain(ctx context.Context, uri string) error {
	fmt.Println("explaining post: ", uri)
	return nil
}
