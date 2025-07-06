package runtime

import (
	"time"

	"github.com/rs/zerolog"
)

func (r *Runtime) StartRepoMirror() {
	log := zerolog.Ctx(r.Ctx).With().Str("module", "repo-mirror").Logger()
	for {
		select {
		case <-r.Ctx.Done():
			log.Info().Msgf("Repo sync stopped")
			return
		default:
			if err := r.syncAccounts(); err != nil {
				if r.Ctx.Err() == nil {
					log.Error().Err(err).Msgf("Failed while sync'n accounts: %s", err)
				}
			} else {
				now := time.Now()
				r.acctMutex.Lock()
				r.lastAccountId = 1 // this needs to be returned, also time
				r.lastAccountTimestamp = now
				r.acctMutex.Unlock()
			}
			time.Sleep(10 * time.Second) // should be able to set this based on the oldest sync timestamp
		}
	}
}

// looks for unfilled or outdated accounts and updates them
func (r *Runtime) syncAccounts() error {
	// first look for empty columns

	// second look for older than some timeframe

	return nil
}

// syncs one account
func (r *Runtime) syncAccount(id uint, did, pds string) error {

	// update some account rate limiting per PDS? or do we use the Bsky public API?

	return nil
}
