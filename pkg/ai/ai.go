package ai

import (
	"context"

	"github.com/blebbit/at-mirror/pkg/runtime"
	"github.com/rs/zerolog/log"
)

type AI struct {
	log log.Logger
	r   *runtime.Runtime
}

func NewAI() (*AI, error) {
	r, err := runtime.NewRuntime(context.Background())
	if err != nil {
		return nil, err
	}

	return &AI{
		log: log.With().Str("module", "ai").Logger(),
		r:   r,
	}, nil
}
