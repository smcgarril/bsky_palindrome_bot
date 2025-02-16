package main

import (
	"context"
	"log/slog"
	"os"

	api "github.com/smcgarril/bsky_palindrome_bot/api"

	"github.com/joho/godotenv"
)

func main() {
	slog.Info("Starting application")

	err := godotenv.Load()
	if err != nil {
		slog.Warn("Warning: No .env file found")
	}

	handle := os.Getenv("HANDLE")
	apikey := os.Getenv("APIKEY")
	server := "https://bsky.social"

	ctx := context.Background()
	if err := api.StartFirehose(ctx, server, handle, apikey); err != nil {
		slog.Error("Firehose encountered an error", "error", err)
	}
}
