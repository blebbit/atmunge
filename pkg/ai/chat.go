package ai

import (
	"context"

	"github.com/blebbit/at-mirror/pkg/ai/ollama"
)

func (a *AI) Chat(ctx context.Context, model, prompt string) (string, error) {
	req := &ollama.ChatRequest{
		Model: model,
		Messages: []ollama.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	resp, err := a.Ollama.Chat(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Message.Content, nil
}
