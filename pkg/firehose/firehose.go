package firehose

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/blebbit/atmunge/pkg/runtime"
	apibsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/jetstream/pkg/client"
	"github.com/bluesky-social/jetstream/pkg/client/schedulers/sequential"
	"github.com/bluesky-social/jetstream/pkg/models"
)

type FirehoseClient struct {
	*runtime.Runtime
	Client *client.Client
	Logger *slog.Logger

	mx        sync.RWMutex
	restarts  int64
	cursor    int64
	seenSeqs  map[int64]struct{}
	highwater int64
	logger    *slog.Logger
}

func NewFirehoseClient(r *runtime.Runtime) (*FirehoseClient, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))

	config := client.DefaultClientConfig()

	relayHost := r.Cfg.RelayHost
	if !strings.HasPrefix(relayHost, "ws") {
		relayHost = "wss://" + relayHost
	}
	// Assuming the jetstream endpoint is at /subscribe
	config.WebsocketURL = relayHost + "/subscribe"

	config.Compress = true

	fc := &FirehoseClient{
		Runtime:  r,
		Logger:   logger,
		seenSeqs: make(map[int64]struct{}),
		logger:   logger,
		cursor:   time.Now().Add(5 * -time.Minute).UnixMicro(), // start 5 minutes ago, should be configurable, or maybe go to database?
	}

	scheduler := sequential.NewScheduler("jetstream_localdev", logger, fc.HandleEvent)

	var err error
	fc.Client, err = client.NewClient(config, logger, scheduler)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return fc, nil
}

func (fc *FirehoseClient) ConnectAndRead(ctx context.Context) error {
	// Every 5 seconds print the events read and bytes read and average event size
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				eventsRead := fc.Client.EventsRead.Load()
				bytesRead := fc.Client.BytesRead.Load()
				avgEventSize := bytesRead / max(1, eventsRead)
				fc.Logger.Info("stats", "events_read", eventsRead, "bytes_read", bytesRead, "avg_event_size", avgEventSize, "cursor", time.UnixMicro(fc.cursor).Local().Format("15:04:05"))
			}
		}
	}()

	for {
		start := time.UnixMicro(fc.cursor).Add(-5 * time.Second).UnixMicro()
		fc.Logger.Info("connect", "start", time.UnixMicro(start).Local().Format("15:04:05"), "restarts", fc.restarts)
		if err := fc.Client.ConnectAndRead(ctx, &start); err != nil {
			fc.Logger.Error("disconnect", "err", err)
			fc.restarts += 1
		}
	}

	fc.Logger.Info("shutdown")
	return nil
}

func (fc *FirehoseClient) HandleEvent(ctx context.Context, event *models.Event) error {
	fc.mx.Lock()
	defer fc.mx.Unlock()
	fc.cursor = event.TimeUS

	// Unmarshal the record if there is one
	if event.Commit != nil && (event.Commit.Operation == models.CommitOperationCreate || event.Commit.Operation == models.CommitOperationUpdate) {
		switch event.Commit.Collection {
		case "app.bsky.feed.post":
			var post apibsky.FeedPost
			if err := json.Unmarshal(event.Commit.Record, &post); err != nil {
				return fmt.Errorf("failed to unmarshal post: %w", err)
			}
			// h.logger.Info("post", "did", event.Did, "text", post.Text, "time", time.UnixMicro(event.TimeUS).Local().Format("15:04:05"))
		}
	}

	return nil
}
