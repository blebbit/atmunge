package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/blebbit/at-mirror/pkg/ai/ollama"
	"github.com/blebbit/at-mirror/pkg/repo"
	"github.com/bluesky-social/indigo/atproto/syntax"
)

// Topics gets topics for a post
func (a *AI) Topics(ctx context.Context, model, prompt, uri string) error {
	atURI, err := syntax.ParseATURI(uri)
	if err != nil {
		return fmt.Errorf("failed to parse at uri: %w", err)
	}

	dbPath := filepath.Join(a.r.Cfg.RepoDataDir, atURI.Authority().String()+".duckdb")
	recordJSON, err := repo.GetRecord(dbPath, atURI.Collection().String(), atURI.RecordKey().String())
	if err != nil {
		return fmt.Errorf("failed to get record from duckdb: %w", err)
	}

	var record map[string]interface{}
	if err := json.Unmarshal(recordJSON, &record); err != nil {
		return fmt.Errorf("failed to unmarshal record json: %w", err)
	}

	finalPrompt := fmt.Sprintf("What are the topics of the following JSON record? Respond with a simple JSON object with a single key, \"topics\", and a list of strings. For example: {\"topics\": [\"one\", \"two\"]}. JSON record:\n\n%s", recordJSON)
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
