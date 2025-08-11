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

// Embed embeds a post
func (a *AI) Embed(ctx context.Context, model, prompt, uri string) error {
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

	var text string
	if textVal, ok := record["text"]; ok {
		if t, ok := textVal.(string); ok {
			text = t
		}
	}

	if text == "" {
		// Fallback to marshaling the whole record if "text" is not found
		text = string(recordJSON)
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
