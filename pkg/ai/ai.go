package ai

import (
	"context"
	"net/http"

	"github.com/blebbit/at-mirror/pkg/ai/ollama"
	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type AI struct {
	log    zerolog.Logger
	r      *runtime.Runtime
	Ollama *ollama.Client
}

func NewAI() (*AI, error) {
	r, err := runtime.NewRuntime(context.Background())
	if err != nil {
		return nil, err
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	return &AI{
		log:    log.With().Str("module", "ai").Logger(),
		r:      r,
		Ollama: ollama.NewClient(cfg.OllamaHost, http.DefaultClient),
	}, nil
}

