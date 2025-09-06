package runtime

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/blebbit/atmunge/pkg/config"
	"github.com/blebbit/atmunge/pkg/db"
	"github.com/blebbit/atmunge/pkg/rlproxy"
)

type Runtime struct {
	// shared resources
	Ctx    context.Context
	Cfg    *config.Config
	DB     *gorm.DB
	Proxy  *rlproxy.Proxy
	Client *http.Client

	// PLC mirror fields
	MaxDelay            time.Duration
	limiter             *rate.Limiter
	plcMutex            sync.RWMutex
	lastRecordTimestamp time.Time

	// Account sync fields
	acctMutex            sync.RWMutex
	lastAccountId        int
	lastAccountTimestamp time.Time
}

func NewRuntime(ctx context.Context) (*Runtime, error) {
	appCfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	r := &Runtime{
		Ctx:      ctx,
		Cfg:      appCfg,
		Proxy:    rlproxy.New(client),
		Client:   client,
		limiter:  rate.NewLimiter(plcRateLimit, 4),
		MaxDelay: plcMaxDelay,
	}

	if r.Cfg.DBUrl != "" {
		// db setup
		DB, err := db.GetClient(r.Cfg.DBUrl, ctx)
		if err != nil {
			return nil, err
		}
		r.DB = DB
	}

	return r, nil
}

func (r *Runtime) LastRecordTimestamp(ctx context.Context) (time.Time, error) {
	r.plcMutex.RLock()
	t := r.lastRecordTimestamp
	r.plcMutex.RUnlock()
	if !t.IsZero() {
		return t, nil
	}

	ts := ""
	err := r.DB.WithContext(ctx).Model(&db.PLCLogEntry{}).Select("plc_timestamp").Order("plc_timestamp desc").Limit(1).Take(&ts).Error
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
