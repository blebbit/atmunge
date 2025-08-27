package ai

import (
	"context"
	"fmt"

	"github.com/blebbit/atmunge/pkg/ai/ollama"
)

const defaultTopicsSystemPrompt = `You are an expert at identifying topics in social media posts.
The user will provide you with the content of a post, and you will identify its topics.
The post content is a JSON object.
Respond with a simple JSON object with a single key, "topics", and a list of strings.
For example: {"topics": ["one", "two"]}.`

// Topics gets topics for a post
func (a *AI) Topics(ctx context.Context, model, prompt, content string) error {
	if prompt == "" {
		prompt = defaultTopicsSystemPrompt
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
