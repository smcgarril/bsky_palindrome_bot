package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/smcgarril/bsky_palindrome_bot/internal/bot"
	"github.com/smcgarril/bsky_palindrome_bot/internal/server"

	"github.com/joho/godotenv"
)

const (
	blueskyServer = "https://bsky.social"
	firehoseURL   = "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos"
)

func main() {
	slog.Info("Starting application")

	err := godotenv.Load()
	if err != nil {
		slog.Warn("Warning: No .env file found")
	}

	apikey := os.Getenv("APIKEY")
	handle := os.Getenv("HANDLE")

	b := bot.NewBot(firehoseURL, blueskyServer, handle, apikey)

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Received shutdown signal")
		cancel()
		b.Stop()
		os.Exit(0)
	}()

	if err := b.Start(ctx); err != nil {
		slog.Error("Error starting bot", "error", err)
		os.Exit(1)
	}

	go func() {
		http.HandleFunc("/health", server.HealthCheck)
		slog.Info("Starting health check server on :8080")

		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("Failed to start health check server:", err)
		}
	}()

	slog.Info("Application has stopped")
}

// package main

// import (
// 	"context"
// 	"log"
// 	"log/slog"
// 	"net/http"
// 	"os"

// 	api "github.com/smcgarril/bsky_palindrome_bot/api"

// 	"github.com/joho/godotenv"
// )

// func main() {
// 	slog.Info("Starting application")

// 	err := godotenv.Load()
// 	if err != nil {
// 		slog.Warn("Warning: No .env file found")
// 	}

// 	handle := os.Getenv("HANDLE")
// 	apikey := os.Getenv("APIKEY")
// 	server := "https://bsky.social"

// 	slog.Info("Starting firehose", "server", server, "handle", handle)

// 	go func() {
// 		http.HandleFunc("/health", api.HealthCheck)
// 		slog.Info("Starting health check server on :8080")

// 		if err := http.ListenAndServe(":8080", nil); err != nil {
// 			log.Fatal("Failed to start health check server:", err)
// 		}
// 	}()

// 	ctx := context.Background()
// 	if err := api.StartFirehose(ctx, server, handle, apikey); err != nil {
// 		slog.Error("Firehose encountered an error", "error", err)
// 	}
// }
