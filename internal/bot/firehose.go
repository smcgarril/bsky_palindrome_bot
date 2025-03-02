package bot

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"
	"github.com/gorilla/websocket"
)

// type Event = *atproto.SyncSubscribeRepos_Commit

type FirehoseClient interface {
	Connect(ctx context.Context) error
	HandleRepoCommit(ctx context.Context, evt *atproto.SyncSubscribeRepos_Commit) error
	Close()
}

type Firehose struct {
	URL       string
	Conn      *websocket.Conn
	Scheduler *sequential.Scheduler
	Logger    *slog.Logger
	Handler   func(ctx context.Context, evt *atproto.SyncSubscribeRepos_Commit) error
	Bot       *Bot
}

func NewFirehose(url string, handler func(ctx context.Context, evt *atproto.SyncSubscribeRepos_Commit) error, logger *slog.Logger, bot *Bot) *Firehose {
	return &Firehose{
		URL:     url,
		Handler: handler,
		Logger:  logger,
		Bot:     bot,
	}
}

func (f *Firehose) Connect(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.Dial(f.URL, http.Header{})
	f.Logger.Info("Started websocket")
	if err != nil {
		return fmt.Errorf("failed to connect to firehose: %w", err)
	}

	// Connect to websocket
	f.Conn = conn
	f.Logger.Info("Connected to Bluesky firehose", "url", f.URL)

	// Define event handler
	rsc := &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			return f.Handler(ctx, evt)
		},
	}

	// Initialize event scheduler
	f.Scheduler = sequential.NewScheduler("firehose", rsc.EventHandler)

	// Start processing stream
	events.HandleRepoStream(ctx, conn, f.Scheduler, f.Logger)

	return nil
}

func (f *Firehose) Close() {
	if f.Conn != nil {
		f.Conn.Close()
		f.Logger.Info("Closed firehose connection")
	}
}

func (f *Firehose) HandleRepoCommit(ctx context.Context, evt *atproto.SyncSubscribeRepos_Commit) error {
	rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
	if err != nil {
		slog.Error("Error reading repo", "error", err)
		return nil
	}

	for _, op := range evt.Ops {
		if isCreateOrUpdate(op.Action) {
			r, err := NewRecord(ctx, *rr, evt, op)
			if err != nil {
				slog.Warn("Error creating record", "error", err)
				continue
			}
			err = r.ProcessRecord(ctx, f.Bot.Server, f.Bot.Handle, f.Bot.APIKey)
			if err != nil {
				slog.Warn("Error processing record", "error", err)
			}
		}
	}
	return nil
}

func isCreateOrUpdate(action string) bool {
	ek := repomgr.EventKind(action)
	return ek == repomgr.EvtKindCreateRecord || ek == repomgr.EvtKindUpdateRecord
}
