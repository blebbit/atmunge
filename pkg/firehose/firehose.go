package firehose

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
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

	h := &handler{
		seenSeqs: make(map[int64]struct{}),
		logger:   logger,
	}

	scheduler := sequential.NewScheduler("jetstream_localdev", logger, h.HandleEvent)

	c, err := client.NewClient(config, logger, scheduler)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &FirehoseClient{
		Runtime: r,
		Client:  c,
		Logger:  logger,
	}, nil
}

func (fc *FirehoseClient) ConnectAndRead(ctx context.Context) error {
	cursor := time.Now().Add(5 * -time.Minute).UnixMicro()

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
				fc.Logger.Info("stats", "events_read", eventsRead, "bytes_read", bytesRead, "avg_event_size", avgEventSize)
			}
		}
	}()

	if err := fc.Client.ConnectAndRead(ctx, &cursor); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	fc.Logger.Info("shutdown")
	return nil
}

type handler struct {
	seenSeqs  map[int64]struct{}
	highwater int64
	logger    *slog.Logger
}

func (h *handler) HandleEvent(ctx context.Context, event *models.Event) error {
	// Unmarshal the record if there is one
	if event.Commit != nil && (event.Commit.Operation == models.CommitOperationCreate || event.Commit.Operation == models.CommitOperationUpdate) {
		switch event.Commit.Collection {
		case "app.bsky.feed.post":
			var post apibsky.FeedPost
			if err := json.Unmarshal(event.Commit.Record, &post); err != nil {
				return fmt.Errorf("failed to unmarshal post: %w", err)
			}
			h.logger.Info("post", "did", event.Did, "text", post.Text, "time", time.UnixMicro(event.TimeUS).Local().Format("15:04:05"))
		}
	}

	return nil
}
