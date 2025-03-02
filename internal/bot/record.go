package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bluesky-social/indigo/api/atproto"
	apibsky "github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	gobot "github.com/danrusei/gobot-bsky"
)

type Record struct {
	Repo         repo.Repo
	Event        *atproto.SyncSubscribeRepos_Commit
	Op           *atproto.SyncSubscribeRepos_RepoOp
	CID          util.LexLink
	Post         apibsky.FeedPost
	PostID       string
	AuthorDID    string
	AuthorHandle string
	AuthorLink   string
}

func NewRecord(ctx context.Context, rr repo.Repo, evt *atproto.SyncSubscribeRepos_Commit, op *atproto.SyncSubscribeRepos_RepoOp) (*Record, error) {
	rc, rec, err := rr.GetRecord(ctx, op.Path)
	if err != nil {
		return nil, fmt.Errorf("error getting record %s: %w", op.Path, err)
	}

	if util.LexLink(rc) != *op.Cid {
		return nil, fmt.Errorf("mismatch in record and op cid: %s != %s", rc, *op.Cid)
	}

	lex := util.LexiconTypeDecoder{Val: rec}
	var pst apibsky.FeedPost

	b, err := lex.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %w", err)
	}

	err = json.Unmarshal(b, &pst)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling post: %w", err)
	}

	authorDID := evt.Repo
	postID := extractPostID(op.Path)

	// Don't retreive AuthorLink unless Record will be processed for palindrome
	return &Record{
		Repo:      rr,
		Event:     evt,
		Op:        op,
		CID:       *op.Cid,
		Post:      pst,
		PostID:    postID,
		AuthorDID: authorDID,
	}, nil

}

func (r *Record) isValidPost() bool {
	return r.Post.LexiconTypeID == "app.bsky.feed.post" && len(r.Post.Langs) > 0 && r.Post.Langs[0] == "en"
}

func (r *Record) ProcessRecord(ctx context.Context, server, handle, apiKey string) error {
	if r.isValidPost() && len(r.Post.Text) > 6 {
		if palindrome, ok := Palindrome(r.Post.Text); ok {
			slog.Info("Palindrome found", "text", palindrome)

			authorHandle, err := r.getAuthorHandle(r.AuthorDID)
			if err != nil {
				slog.Error("Error fetching handle for DID", "error", err)
				authorHandle = r.AuthorDID
			}
			r.AuthorHandle = authorHandle

			postLink := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", authorHandle, r.PostID)
			slog.Info("Post by", "handle", authorHandle, "did", r.AuthorDID)
			slog.Info("Post link", "link", postLink)

			agent := gobot.NewAgent(ctx, server, handle, apiKey)

			p, err := NewPost(r.AuthorDID, r.AuthorHandle, r.PostID, palindrome)
			if err != nil {
				return fmt.Errorf("error creating new post object: %w", err)
			}

			err = p.DoPost(ctx, agent, r.AuthorHandle, r.AuthorDID)
			if err != nil {
				slog.Error("Error posting palindrome", "error", err)
			}
		}

	}
	return nil
}

func (r *Record) getAuthorHandle(did string) (string, error) {
	url := "https://plc.directory/" + did
	slog.Info("Fetching DID info", "url", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch DID info: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AlsoKnownAs []string `json:"alsoKnownAs"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to parse DID response: %w", err)
	}

	if len(result.AlsoKnownAs) == 0 {
		return "", fmt.Errorf("no known handles found for DID: %s", did)
	}

	// Extract handle from "at://handle"
	handle := result.AlsoKnownAs[0]
	if strings.HasPrefix(handle, "at://") {
		handle = strings.TrimPrefix(handle, "at://")
	}

	return handle, nil
}

func extractPostID(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx != -1 {
		return path[idx+1:]
	}
	return ""
}
