package ai

import (
	"context"
	"fmt"

	"github.com/blebbit/atmunge/pkg/ai/ollama"
)

const defaultSafetySystemPrompt = `You are an expert at assessing the safety of social media posts.
The user will provide you with the content of a post, and you will assess its safety.
The post content is a JSON object.
Look for harmful, unethical, racist, sexist, toxic, dangerous, or illegal content.
Respond with a simple JSON object with a single key, "safe", and a boolean value.
For example: {"safe": true}.`

// Safety gets safety status for a post
func (a *AI) Safety(ctx context.Context, model, prompt, content string) error {
	if prompt == "" {
		prompt = defaultSafetySystemPrompt
	}

	resp, err := a.Ollama.Generate(ctx, &ollama.GenerateRequest{
		Model:  model,
		Prompt: content,
		System: prompt,
	})
	if err != nil {
		return fmt.Errorf("failed to get completion from ollama: %w", err)
	}

	fmt.Println(resp.Response)

	return nil
}
