package runtime

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"gorm.io/gorm"

	plcdb "github.com/blebbit/plc-mirror/pkg/db"
)

const (
	// plc.directory settings
	// Current rate limit is `500 per five minutes`, lets stay a bit under it.
	defaultRateLimit  = rate.Limit(450.0 / 300)
	caughtUpRateLimit = rate.Limit(0.2)
	caughtUpThreshold = 10 * time.Minute
	maxDelay          = 5 * time.Minute
)

type Runtime struct {
	ctx context.Context
	db  *gorm.DB

	MaxDelay time.Duration

	upstream *url.URL
	limiter  *rate.Limiter

	mu                      sync.RWMutex
	lastCompletionTimestamp time.Time
	lastRecordTimestamp     time.Time

	handler http.HandlerFunc
}

func NewRuntime(ctx context.Context, db *gorm.DB) (*Runtime, error) {
	u, err := url.Parse("https://plc.directory")
	if err != nil {
		return nil, err
	}
	u.Path, err = url.JoinPath(u.Path, "export")
	if err != nil {
		return nil, err
	}

	r := &Runtime{
		ctx:      ctx,
		db:       db,
		upstream: u,
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastCompletionTimestamp
}

func (r *Runtime) LastRecordTimestamp(ctx context.Context) (time.Time, error) {
	r.mu.RLock()
	t := r.lastRecordTimestamp
	r.mu.RUnlock()
	if !t.IsZero() {
		return t, nil
	}

	ts := ""
	err := r.db.WithContext(ctx).Model(&plcdb.PLCLogEntry{}).Select("plc_timestamp").Order("plc_timestamp desc").Limit(1).Take(&ts).Error
	if err != nil {
		return time.Time{}, err
	}
	dbTimestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing timestamp %q: %w", ts, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.lastRecordTimestamp.IsZero() {
		r.lastRecordTimestamp = dbTimestamp
	}
	if r.lastRecordTimestamp.After(dbTimestamp) {
		return r.lastRecordTimestamp, nil
	}
	return dbTimestamp, nil
}
