package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"log/slog"
	"os"

	"github.com/bluesky-social/indigo/api/atproto"
	apibsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"

	"github.com/gorilla/websocket"
)

const firehoseURI = "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos"

var (
	eventCount    uint64
	postCount     uint64
	longPostCount uint64

	totalEvents    uint64
	totalPosts     uint64
	totalLongPosts uint64

	minuteEventCounts    []uint64
	minutePostCounts     []uint64
	minuteLongPostCounts []uint64
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func StartFirehose(ctx context.Context, server, handle, apikey string) error {
	// Firehose connection setup...
	ticker := time.NewTicker(1 * time.Second)
	minuteTicker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	defer minuteTicker.Stop()

	go func() {
		for range ticker.C {
			events := atomic.SwapUint64(&eventCount, 0)
			posts := atomic.SwapUint64(&postCount, 0)
			longPosts := atomic.SwapUint64(&longPostCount, 0)

			atomic.AddUint64(&totalEvents, events)
			atomic.AddUint64(&totalPosts, posts)
			atomic.AddUint64(&totalLongPosts, longPosts)

			slog.Info("Events per second", "total", events, "posts", posts, "long posts", longPosts)
		}
	}()

	// // Collect minute-based metrics for averages and spikes
	// go func() {
	// 	for range minuteTicker.C {
	// 		events := atomic.LoadUint64(&totalEvents)
	// 		posts := atomic.LoadUint64(&totalPosts)
	// 		longPosts := atomic.LoadUint64(&totalLongPosts)

	// 		// Track minute-level counts
	// 		minuteEventCount := atomic.SwapUint64(&eventCount, 0)
	// 		minutePostCount := atomic.SwapUint64(&postCount, 0)
	// 		minuteLongPostCount := atomic.SwapUint64(&longPostCount, 0)

	// 		minuteEventCounts = append(minuteEventCounts, minuteEventCount)
	// 		minutePostCounts = append(minutePostCounts, minutePostCount)
	// 		minuteLongPostCounts = append(minuteLongPostCounts, minuteLongPostCount)

	// 		// Calculate averages
	// 		avgEvents := calculateAverage(minuteEventCounts)
	// 		avgPosts := calculateAverage(minutePostCounts)
	// 		avgLongPosts := calculateAverage(minuteLongPostCounts)

	// 		// Detect spikes (e.g., 2x the moving average)
	// 		if minuteEventCount > (2 * uint64(avgEvents)) {
	// 			slog.Warn("Spike detected in event count!", "current", minuteEventCount, "average", avgEvents)
	// 		}
	// 		if minutePostCount > (2 * uint64(avgPosts)) {
	// 			slog.Warn("Spike detected in post count!", "current", minutePostCount, "average", avgPosts)
	// 		}
	// 		if minuteLongPostCount > (2 * uint64(avgLongPosts)) {
	// 			slog.Warn("Spike detected in long post count!", "current", minuteLongPostCount, "average", avgLongPosts)
	// 		}

	// 		slog.Info("Minute Summary",
	// 			"total_events", events,
	// 			"total_posts", posts,
	// 			"total_long_posts", longPosts,
	// 			"average_events_per_minute", avgEvents,
	// 			"average_posts_per_minute", avgPosts,
	// 			"average_long_posts_per_minute", avgLongPosts)
	// 	}
	// }()

	// Connect to the WebSocket
	con, _, err := websocket.DefaultDialer.Dial(firehoseURI, http.Header{})
	if err != nil {
		atomic.AddUint64(&eventCount, 1)
		return fmt.Errorf("failed to connect to firehose: %w", err)
	}
	defer con.Close()

	rsc := &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			atomic.AddUint64(&eventCount, 1)
			return handleRepoCommit(ctx, evt, server, handle, apikey)
		},
	}

	sched := sequential.NewScheduler("myfirehose", rsc.EventHandler)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	events.HandleRepoStream(ctx, con, sched, logger)
	return nil
}

func handleRepoCommit(ctx context.Context, evt *atproto.SyncSubscribeRepos_Commit, server, handle, apikey string) error {
	rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
	if err != nil {
		slog.Error("Error reading repor", "error", err)
		return nil
	}

	for _, op := range evt.Ops {
		if isCreateOrUpdate(op.Action) {
			// err := processRecord(ctx, *rr, evt, op, server, handle, apikey)
			_ = processRecord(ctx, *rr, evt, op, server, handle, apikey)
			// if err != nil {
			// 	slog.Warn("Error processing record", "error", err)
			// }
		}
	}
	return nil
}

func isCreateOrUpdate(action string) bool {
	ek := repomgr.EventKind(action)
	return ek == repomgr.EvtKindCreateRecord || ek == repomgr.EvtKindUpdateRecord
}

func processRecord(ctx context.Context, rr repo.Repo, evt *atproto.SyncSubscribeRepos_Commit, op *atproto.SyncSubscribeRepos_RepoOp, server, handle, apikey string) error {
	rc, rec, err := rr.GetRecord(ctx, op.Path)
	if err != nil {
		return fmt.Errorf("error getting record %s: %w", op.Path, err)
	}

	if lexutil.LexLink(rc) != *op.Cid {
		return fmt.Errorf("mismatch in record and op cid: %s != %s", rc, *op.Cid)
	}

	lex := lexutil.LexiconTypeDecoder{Val: rec}
	var pst apibsky.FeedPost

	b, err := lex.MarshalJSON()
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	err = json.Unmarshal(b, &pst)
	if err != nil {
		return fmt.Errorf("error unmarshalling post: %w", err)
	}

	if pst.LexiconTypeID == "app.bsky.feed.post" && len(pst.Langs) > 0 && pst.Langs[0] == "en" {
		atomic.AddUint64(&postCount, 1)
		if len(pst.Text) > 6 {
			atomic.AddUint64(&longPostCount, 1)
			if palindrome, ok := Palindrome(pst.Text); ok {

				slog.Info("Palindrome found", "text", palindrome)

				authorDID := evt.Repo
				postID := ExtractPostID(op.Path)

				authorHandle, err := GetAuthorHandle(authorDID)
				if err != nil {
					slog.Error("Error fetching handle for DID", "error", err)
					authorHandle = authorDID
				}

				postLink := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", authorHandle, postID)

				slog.Info("Post by", "handle", authorHandle, "did", authorDID)
				slog.Info("Post link", "link", postLink)

				err = Post(ctx, palindrome, server, handle, authorHandle, authorDID, postID, apikey)
				if err != nil {
					slog.Error("Error posting palindrome", "error", err)
				}
			}

		}
	}
	return nil
}

func calculateAverage(data []uint64) float64 {
	if len(data) == 0 {
		return 0
	}
	var sum uint64
	for _, count := range data {
		sum += count
	}
	return float64(sum) / float64(len(data))
}
