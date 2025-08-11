package ai

import (
	"context"
	"fmt"
)

func (a *AI) Hack(ctx context.Context, uri string) error {
	fmt.Println("hacking post: ", uri)
	return nil
}
