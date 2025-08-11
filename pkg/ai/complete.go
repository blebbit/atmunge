package ai

import (
	"context"

	"github.com/blebbit/at-mirror/pkg/ai/ollama"
)

func (a *AI) Complete(ctx context.Context, model, prompt string) (string, error) {
	req := &ollama.GenerateRequest{
		Model:  model,
		Prompt: prompt,
	}

	resp, err := a.Ollama.Generate(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Response, nil
}
