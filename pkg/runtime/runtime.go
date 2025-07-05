package runtime

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"gorm.io/gorm"

	// "github.com/blebbit/at-mirror/pkg/config"
	plcdb "github.com/blebbit/at-mirror/pkg/db"
)

const (
	// plc.directory settings
	// Current rate limit is `500 per five minutes`, lets stay a bit under it.
	defaultRateLimit  = rate.Limit(450.0 / 300.0)
	caughtUpRateLimit = rate.Limit(0.2)
	caughtUpThreshold = 10 * time.Minute
	maxDelay          = 5 * time.Minute
)

type Runtime struct {
	// shared resources
	Ctx context.Context
	// Cfg *config.Config
	DB *gorm.DB

	// PLC mirror fields
	upstream                *url.URL
	MaxDelay                time.Duration
	limiter                 *rate.Limiter
	plcMutex                sync.RWMutex
	lastCompletionTimestamp time.Time
	lastRecordTimestamp     time.Time

	// Account sync fields
	acctMutex            sync.RWMutex
	lastAccountId        int
	lastAccountTimestamp time.Time
}

func NewRuntime(ctx context.Context, db *gorm.DB) (*Runtime, error) {

	r := &Runtime{
		Ctx:      ctx,
		DB:       db,
		upstream: plcUrl(),
		limiter:  rate.NewLimiter(defaultRateLimit, 4),
		MaxDelay: maxDelay,
	}

	return r, nil
}

func (r *Runtime) updateRateLimit(lastRecordTimestamp time.Time) {
	// Reduce rate limit if we are caught up, to get new records in larger batches.
	desiredRate := defaultRateLimit
	if time.Since(lastRecordTimestamp) < caughtUpThreshold {
		desiredRate = caughtUpRateLimit
	}
	if math.Abs(float64(r.limiter.Limit()-desiredRate)) > 0.0000001 {
		r.limiter.SetLimit(rate.Limit(desiredRate))
	}
}

func (r *Runtime) LastCompletion() time.Time {
	r.plcMutex.RLock()
	defer r.plcMutex.RUnlock()
	return r.lastCompletionTimestamp
}

func (r *Runtime) LastRecordTimestamp(ctx context.Context) (time.Time, error) {
	r.plcMutex.RLock()
	t := r.lastRecordTimestamp
	r.plcMutex.RUnlock()
	if !t.IsZero() {
		return t, nil
	}

	ts := ""
	err := r.DB.WithContext(ctx).Model(&plcdb.PLCLogEntry{}).Select("plc_timestamp").Order("plc_timestamp desc").Limit(1).Take(&ts).Error
	if err != nil {
		return time.Time{}, err
	}
	dbTimestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing timestamp %q: %w", ts, err)
	}

	r.plcMutex.Lock()
	defer r.plcMutex.Unlock()
	if r.lastRecordTimestamp.IsZero() {
		r.lastRecordTimestamp = dbTimestamp
	}
	if r.lastRecordTimestamp.After(dbTimestamp) {
		return r.lastRecordTimestamp, nil
	}
	return dbTimestamp, nil
}
