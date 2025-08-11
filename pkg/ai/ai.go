package ai

import (
	"github.com/rs/zerolog/log"
)

type AI struct {
	log log.Logger
}

func NewAI() (*AI, error) {
	return &AI{
		log: log.With().Str("module", "ai").Logger(),
	}, nil
}
