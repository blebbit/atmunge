package ai

import (
	"context"

	"github.com/blebbit/atmunge/pkg/ai/ollama"
)

func (a *AI) Chat(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
	finalPrompt := userPrompt
	if systemPrompt != "" {
		finalPrompt = systemPrompt + "\n\n" + userPrompt
	}
	req := &ollama.ChatRequest{
		Model: model,
		Messages: []ollama.Message{
			{
				Role:    "user",
				Content: finalPrompt,
			},
		},
	}

	resp, err := a.Ollama.Chat(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Message.Content, nil
}
