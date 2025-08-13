package runtime

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/blebbit/at-mirror/pkg/config"
	"github.com/blebbit/at-mirror/pkg/db"
	plcdb "github.com/blebbit/at-mirror/pkg/db"
	"github.com/blebbit/at-mirror/pkg/rlproxy"
)

type Runtime struct {
	// shared resources
	Ctx    context.Context
	Cfg    *config.Config
	DB     *gorm.DB
	Proxy  *rlproxy.Proxy
	Client *http.Client

	// PLC mirror fields
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

func NewRuntime(ctx context.Context) (*Runtime, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	// db setup
	DB, err := db.GetClient(cfg.DBUrl, ctx)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	r := &Runtime{
		Ctx:      ctx,
		Cfg:      cfg,
		DB:       DB,
		Proxy:    rlproxy.New(client),
		Client:   client,
		limiter:  rate.NewLimiter(plcRateLimit, 4),
		MaxDelay: plcMaxDelay,
	}

	return r, nil
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
