package main

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"

	api "github.com/smcgarril/bsky_palindrome_bot/api"

	"github.com/joho/godotenv"
)

func setupLogging() (*os.File, error) {
	// Open the log file for writing (create if it doesn't exist, append if it does)
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// Create a multi-writer (console + log file)
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Set up the slog handler
	handler := slog.NewTextHandler(multiWriter, nil)
	logger := slog.New(handler)

	// Set the default logger
	slog.SetDefault(logger)

	return logFile, nil
}

func main() {
	logFile, err := setupLogging()
	if err != nil {
		slog.Error("Failed to set up logging", "error", err)
		return
	}
	defer logFile.Close()

	slog.Info("Starting application")

	err = godotenv.Load()
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
			log.Fatal("Failed to start health check server:", err)
		}
	}()

	ctx := context.Background()
	if err := api.StartFirehose(ctx, server, handle, apikey); err != nil {
		slog.Error("Firehose encountered an error", "error", err)
	}
}
