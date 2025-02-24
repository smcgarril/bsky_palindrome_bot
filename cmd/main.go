package main

import (
	"context"
	"log/slog"
	"net/http"
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

	slog.Info("Starting firehose", "server", server, "handle", handle)

	go func() {
		http.HandleFunc("/health", api.HealthCheck)
		slog.Info("Starting health check server on :8080")

		if err := http.ListenAndServe(":8080", nil); err != nil {
			slog.Error("Failed to start health check server", "error", err)
			os.Exit(1)
		}
	}()

	dictPath := "/usr/share/dict/words" // Or use a custom wordlist
	dictionary, err := api.LoadDictionary(dictPath)
	if err != nil {
		slog.Error("Failed to load dictionary:", "error", err)
		os.Exit(1)
	}

	validator := api.NewWordSegmentValidator(dictionary)

	ctx := context.Background()
	if err := api.StartFirehose(ctx, validator, server, handle, apikey); err != nil {
		slog.Error("Firehose encountered an error", "error", err)
	}
}
