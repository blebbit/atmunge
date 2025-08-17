package runtime

import (
	"time"

	"github.com/rs/zerolog"
)

func (r *Runtime) StartPLCMirror() {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "plc").Logger()
	for {
		select {
		case <-r.Ctx.Done():
			log.Info().Msgf("PLC mirror stopped")
			return
		default:
			if err := r.BackfillPlcLogs(); err != nil {
				if r.Ctx.Err() == nil {
					log.Error().Err(err).Msgf("Failed to get new log entries from PLC: %s", err)
				}
			}
			time.Sleep(time.Duration(r.Cfg.PlcMirrorDelay) * time.Second)
		}
	}
}
