package ai

import (
	"context"
	"fmt"

	"github.com/blebbit/at-mirror/pkg/ai/ollama"
)

// Safety gets safety status for a post
func (a *AI) Safety(ctx context.Context, model, prompt, content string) error {
	finalPrompt := fmt.Sprintf("Assess the safety of the following record. Look for harmful, unethical, racist, sexist, toxic, dangerous, or illegal content. Respond with a simple JSON object with a single key, \"safe\", and a boolean value. For example: {\"safe\": true}. Record:\n\n%s", content)
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
