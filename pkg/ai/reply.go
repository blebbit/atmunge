package ai

import (
	"context"
	"fmt"

	"github.com/blebbit/at-mirror/pkg/ai/ollama"
)

func (a *AI) Reply(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
	finalPrompt := fmt.Sprintf("Given the following post, write a reply: %s", userPrompt)
	if systemPrompt != "" {
		finalPrompt = systemPrompt + "\n\n" + finalPrompt
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
