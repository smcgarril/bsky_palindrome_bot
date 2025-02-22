package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
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

type Record struct {
	ctx    context.Context
	rr     *repo.Repo
	evt    *atproto.SyncSubscribeRepos_Commit
	op     *atproto.SyncSubscribeRepos_RepoOp
	server string
	handle string
	apikey string
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func Worker(id int, records <-chan Record, wg *sync.WaitGroup) {
	defer wg.Done()
	for record := range records {
		// for range records {
		// Comment this out for production
		// events := len(records)
		// if events > 50 {
		// 	slog.Info("Worker", "id", id, "processing\n", record, "Number of events:", events)
		// }
		// time.Sleep(10 * time.Millisecond)
		// slog.Info("Worker", "id", id, "processing", j)

		// // actual processing
		err := processRecord(record) // <-- need to pass Record
		if err != nil {
			slog.Warn("Error processing record", "error", err)
		}

	}
}

func StartFirehose(ctx context.Context, eventQueue chan Record, fallBackQueue chan Record, server, handle, apikey string) error {
	// Connect to the WebSocket
	con, _, err := websocket.DefaultDialer.Dial(firehoseURI, http.Header{})
	if err != nil {
		return fmt.Errorf("failed to connect to firehose: %w", err)
	}
	defer con.Close()

	rsc := &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			return handleRepoCommit(ctx, eventQueue, fallBackQueue, evt, server, handle, apikey)
		},
	}

	sched := sequential.NewScheduler("myfirehose", rsc.EventHandler)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	events.HandleRepoStream(ctx, con, sched, logger)
	return nil
}

func handleRepoCommit(ctx context.Context, eventQueue chan Record, fallBackQueue chan Record, evt *atproto.SyncSubscribeRepos_Commit, server, handle, apikey string) error {
	rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
	if err != nil {
		slog.Error("Error reading repor", "error", err)
		return nil
	}

	for _, op := range evt.Ops {
		if isCreateOrUpdate(op.Action) {
			record := Record{ctx, rr, evt, op, server, handle, apikey}

			select {
			case eventQueue <- record:
				// slog.Info("Queue lengths", "Primary queue", len(eventQueue), "Fallback queue", len(fallBackQueue))
			default:
				// slog.Warn("Primary queue full. Record sent to fallback queue", "", record)
				select {
				case fallBackQueue <- record:
					// slog.Info("Queue lengths", "Primary queue", len(eventQueue), "Fallback queue", len(fallBackQueue))
					slog.Warn("Record enqueued to fallback queue", "", record)
				default:
					slog.Error("Fallback queue also full. Dropping record", "", record)
				}
			}

			err := processRecord(record)
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

// refactor to accept Record
func processRecord(record Record) error {
	rc, rec, err := record.rr.GetRecord(record.ctx, record.op.Path)
	if err != nil {
		return fmt.Errorf("error getting record %s: %w", record.op.Path, err)
	}

	if lexutil.LexLink(rc) != *record.op.Cid {
		return fmt.Errorf("mismatch in record and op cid: %s != %s", rc, record.op.Cid)
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
		if len(pst.Text) > 6 {
			if palindrome, ok := Palindrome(pst.Text); ok {
				slog.Info("Palindrome found", "text", palindrome)

				authorDID := record.evt.Repo
				postID := ExtractPostID(record.op.Path)

				authorHandle, err := GetAuthorHandle(authorDID)
				if err != nil {
					slog.Error("Error fetching handle for DID", "error", err)
					authorHandle = authorDID
				}

				postLink := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", authorHandle, postID)

				slog.Info("Post by", "handle", authorHandle, "did", authorDID)
				slog.Info("Post link", "link", postLink)

				err = Post(record.ctx, palindrome, record.server, record.handle, authorHandle, authorDID, postID, record.apikey)
				if err != nil {
					slog.Error("Error posting palindrome", "error", err)
				}
			}

		}
	}
	return nil
}

func ProcessFallbackQueue(eventQueue chan Record, fallBackQueue chan Record) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		// slog.Info("Queue lengths", "Primary queue", len(eventQueue), "Fallback queue", len(fallBackQueue))
		for {
			select {
			case record := <-fallBackQueue:
				select {
				case eventQueue <- record:
					// slog.Info("Moved fallback record back to primary queue", "", record)
				default:
					// slog.Warn("Primary queue still full. Record remains in fallback queue", "", record)
					break
				}
			default:
				break
			}
		}
	}
}
