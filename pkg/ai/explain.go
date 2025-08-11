package ai

import (
	"context"
	"fmt"

	"github.com/blebbit/at-mirror/pkg/ai/ollama"
)

// Explain explains a post
func (a *AI) Explain(ctx context.Context, model, prompt, content string) error {
	finalPrompt := fmt.Sprintf("Explain the following record:\n\n%s", content)
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
