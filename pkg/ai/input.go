package ai

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blebbit/atmunge/pkg/repo"
	"github.com/bluesky-social/indigo/atproto/syntax"
)

// ResolveInput resolves the input string.
// It can be a raw string, "-", an "at://" URI, or a file path.
// If it's "-", it reads from stdin.
// If it's an "at://" URI, it fetches the record.
// If it's a file path that exists, it reads the file.
// Otherwise, it returns the raw string.
func (a *AI) ResolveInput(input string) (string, error) {
	if input == "-" {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		return string(bytes), nil
	}

	if strings.HasPrefix(input, "at://") {
		atURI, err := syntax.ParseATURI(input)
		if err != nil {
			return "", fmt.Errorf("failed to parse at uri: %w", err)
		}
		dbPath := a.GetRepoDataDir() + "/" + atURI.Authority().String() + "/repo.duckdb"
		recordJSON, err := repo.GetRecord(dbPath, atURI.Collection().String(), atURI.RecordKey().String())
		if err != nil {
			return "", fmt.Errorf("failed to get record from duckdb: %w", err)
		}
		return string(recordJSON), nil
	}

	if _, err := os.Stat(input); err == nil {
		bytes, err := os.ReadFile(input)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return string(bytes), nil
	}

	return input, nil
}
