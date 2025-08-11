package ai

import (
	"context"
	"fmt"

	"github.com/blebbit/at-mirror/pkg/ai/ollama"
)

func (a *AI) Reply(ctx context.Context, model, prompt string) (string, error) {
	req := &ollama.GenerateRequest{
		Model:  model,
		Prompt: fmt.Sprintf("Given the following post, write a reply: %s", prompt),
	}

	resp, err := a.Ollama.Generate(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Response, nil
}
