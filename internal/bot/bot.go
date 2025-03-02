package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type BotClient interface {
	Start(ctx context.Context) error
	Close()
}

type Bot struct {
	Server   string
	Handle   string
	APIKey   string
	Firehose *Firehose
	Logger   *slog.Logger
}

func NewBot(firehoseURL, blueskyServer, handle, apiKey string) *Bot {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	bot := &Bot{
		Server: blueskyServer,
		Handle: handle,
		APIKey: apiKey,
		Logger: logger,
	}

	bot.Firehose = &Firehose{
		URL:     firehoseURL,
		Logger:  logger,
		Handler: nil,
		Bot:     bot,
	}

	bot.Firehose.Handler = bot.Firehose.HandleRepoCommit

	return bot
}

func (b *Bot) Start(ctx context.Context) error {
	b.Logger.Info("Starting bot")

	err := b.Firehose.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to start firehose: %w", err)
	}

	return nil
}

func (b *Bot) Stop() {
	b.Logger.Info("Stopping bot")
	b.Firehose.Close()
}
