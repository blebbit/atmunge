package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blebbit/atmunge/pkg/ai/ollama"
)

// Embed embeds a post
func (a *AI) Embed(ctx context.Context, model, prompt, content string) error {
	var record map[string]interface{}
	if err := json.Unmarshal([]byte(content), &record); err != nil {
		// if it's not a json, we just treat it as a raw string
	}

	var text string
	if textVal, ok := record["text"]; ok {
		if t, ok := textVal.(string); ok {
			text = t
		}
	}

	if text == "" {
		// Fallback to marshaling the whole record if "text" is not found
		text = content
	}

	if prompt != "" {
		text = prompt + "\n\n" + text
	}

	resp, err := a.Ollama.GenerateEmbeddings(ctx, &ollama.EmbeddingsRequest{
		Model: model,
		Input: []string{text},
	})
	if err != nil {
		return fmt.Errorf("failed to get embeddings from ollama: %w", err)
	}

	if len(resp.Embeddings) > 0 {
		fmt.Println("ollama response (embedding): ", resp.Embeddings[0])
	} else {
		fmt.Println("ollama response: no embeddings generated")
	}

	return nil
}
