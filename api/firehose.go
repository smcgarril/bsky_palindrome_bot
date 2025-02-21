package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func StartFirehose(ctx context.Context, server, handle, apikey string) error {
	// Connect to the WebSocket
	con, _, err := websocket.DefaultDialer.Dial(firehoseURI, http.Header{})
	if err != nil {
		return fmt.Errorf("failed to connect to firehose: %w", err)
	}
	defer con.Close()

	rsc := &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
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
			err := processRecord(ctx, *rr, evt, op, server, handle, apikey)
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
		if len(pst.Text) > 6 {
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
