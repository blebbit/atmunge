package ai

import (
	"context"

	"github.com/blebbit/at-mirror/pkg/ai/ollama"
)

const defaultReplySystemPrompt = `You are an expert at replying to social media posts.
The user will provide you with the content of a post, and you will reply to it.
The post content is a JSON object.
Your reply should be clear, concise, and easy to understand.
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
Just reply to the post in a single paragraph.`

func (a *AI) Reply(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
	if systemPrompt == "" {
		systemPrompt = defaultReplySystemPrompt
	}
	req := &ollama.GenerateRequest{
		Model:  model,
		Prompt: userPrompt,
		System: systemPrompt,
	}

	resp, err := a.Ollama.Generate(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Response, nil
}
