package ai

import (
	"context"
	"fmt"

	"github.com/blebbit/at-mirror/pkg/ai/ollama"
)

const defaultSummarizeSystemPrompt = `You are an expert at summarizing social media posts.
The user will provide you with the content of a post, and you will summarize it.
The post content is a JSON object with the following structure:
{
  "uri": "at://did:plc:...",
  "cid": "...",
  "author": {
    "did": "did:plc:...",
    "handle": "handle.bsky.social",
    "displayName": "display name",
    "description": "description",
    "avatar": "..."
  },
  "record": {
    "text": "post text",
    "createdAt": "..."
  }
}
Your summary should be clear, concise, and easy to understand.
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
Just summarize the post in a single paragraph.
`

// Summarize summarizes a post
func (a *AI) Summarize(ctx context.Context, model, prompt, input string) error {
	if prompt == "" {
		prompt = defaultSummarizeSystemPrompt
	}

	resp, err := a.Ollama.Generate(ctx, &ollama.GenerateRequest{
		Model:  model,
		Prompt: input,
		System: prompt,
	})
	if err != nil {
		return fmt.Errorf("failed to get completion: %w", err)
	}
	fmt.Println(resp.Response)
	return nil
}
