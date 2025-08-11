package ai

import (
	"context"
	"fmt"

	"github.com/blebbit/at-mirror/pkg/ai/ollama"
)

// Topics gets topics for a post
func (a *AI) Topics(ctx context.Context, model, prompt, content string) error {
	finalPrompt := fmt.Sprintf("What are the topics of the following record? Respond with a simple JSON object with a single key, \"topics\", and a list of strings. For example: {\"topics\": [\"one\", \"two\"]}. Record:\n\n%s", content)
	if prompt != "" {
		finalPrompt = prompt + "\n\n" + finalPrompt
	}

	resp, err := a.Ollama.Generate(ctx, &ollama.GenerateRequest{
		Model:  model,
		Prompt: finalPrompt,
	})
	if err != nil {
		return fmt.Errorf("failed to get completion from ollama: %w", err)
	}

	fmt.Println("ollama response: ", resp.Response)

	return nil
}
