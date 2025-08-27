package ai

import (
	"context"
	"fmt"

	"github.com/blebbit/atmunge/pkg/ai/ollama"
)

const defaultExplainSystemPrompt = `You are an expert at explaining social media posts.
The user will provide you with the content of a post, and you will explain it.
The post content is a JSON object.
Your explanation should be clear, concise, and easy to understand.
Do not repeat the content of the post.
Do not use jargon.
Do not be verbose.
Focus on the most important aspects of the post.
Be objective and unbiased.
Do not add any information that is not present in the post.
Do not ask questions.
Do not say "This post is about..." or "In this post...".
Do not use emojis.
Do not use hashtags.
Do not use markdown.
Do not use any formatting.
Do not use quotes.
Do not use links.
Do not use any other information.
Just explain the post in a single paragraph.`

// Explain explains a post
func (a *AI) Explain(ctx context.Context, model, prompt, content string) error {
	if prompt == "" {
		prompt = defaultExplainSystemPrompt
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
