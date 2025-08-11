package ai

import (
	"context"

	"github.com/blebbit/at-mirror/pkg/ai/ollama"
)

func (a *AI) Complete(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
	finalPrompt := userPrompt
	if systemPrompt != "" {
		finalPrompt = systemPrompt + "\n\n" + userPrompt
	}
	req := &ollama.GenerateRequest{
		Model:  model,
		Prompt: finalPrompt,
	}

	resp, err := a.Ollama.Generate(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Response, nil
}
