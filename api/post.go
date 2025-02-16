package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	gobot "github.com/danrusei/gobot-bsky"
)

func Post(ctx context.Context, palindrome, server, handle, authorHandle, authorDID, postID, apikey string) error {
	agent := gobot.NewAgent(ctx, server, handle, apikey)
	agent.Connect(ctx)

	urlStr := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", authorHandle, postID)

	text := fmt.Sprintf("Palindrome Found!\n\n%s\n\nFrom user %s", palindrome, authorHandle)

	post, err := gobot.NewPostBuilder(text).
		WithFacet(gobot.Facet_Link, urlStr, palindrome).
		WithFacet(gobot.Facet_Mention, authorDID, authorHandle).
		Build()
	if err != nil {
		return fmt.Errorf("Error creating post: %v\n", err)
	}

	cid1, uri1, err := agent.PostToFeed(ctx, post)
	if err != nil {
		return fmt.Errorf("Error posting to feed: %v\n", err)
	}

	slog.Info("Posted to feed", "cid", cid1, "uri", uri1)
	return nil
}

func GetAuthorHandle(did string) (string, error) {
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

// Get substring after last "/"
func ExtractPostID(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx != -1 {
		return path[idx+1:]
	}
	return ""
}
