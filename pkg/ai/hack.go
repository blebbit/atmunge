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

func (a *AI) Hack(ctx context.Context, uri string) error {
	fmt.Println("hacking post: ", uri)

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

	prompt := fmt.Sprintf("Extract the topics from the following JSON record:\n\n%s", recordJSON)

	resp, err := a.Ollama.Generate(ctx, &ollama.GenerateRequest{
		Model:  "gemma3:4b",
		Prompt: prompt,
	})
	if err != nil {
		return fmt.Errorf("failed to get completion from ollama: %w", err)
	}

	fmt.Println("ollama response: ", resp.Response)

	return nil
}
