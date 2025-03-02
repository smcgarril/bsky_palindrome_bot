package bot

import (
	"context"
	"fmt"
	"log/slog"

	gobot "github.com/danrusei/gobot-bsky"
)

type Post struct {
	Palindrome string
	PostText   string
	PostURL    string
}

func NewPost(authorDID, authorHandle, postID, palindrome string) (*Post, error) {
	postText := fmt.Sprintf("Palindrome Found!\n\n%s\n\nFrom user %s", palindrome, authorHandle)
	postURL := fmt.Sprintf("https://bsky.app/profile/%s/post/%s", authorHandle, postID)

	newPost := &Post{
		Palindrome: palindrome,
		PostText:   postText,
		PostURL:    postURL,
	}

	return newPost, nil
}

func (p *Post) DoPost(ctx context.Context, agent gobot.BskyAgent, authorHandle, authorDID string) error {
	agent.Connect(ctx)

	post, err := gobot.NewPostBuilder(p.PostText).
		WithFacet(gobot.Facet_Link, p.PostURL, p.Palindrome).
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
