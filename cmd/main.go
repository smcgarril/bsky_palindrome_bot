package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	api "github.com/smcgarril/bsky_palindrome_bot/api"

	"github.com/joho/godotenv"
)

func main() {
	// Benchmark testing
	benchMode := flag.Bool("bench", false, "Run benchmark instead of main application")
	flag.Parse()

	if *benchMode {
		slog.Info("Running benchmark mode")
		api.RunBenchmark()
		return
	}

	// Production mode
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
			log.Fatal("Failed to start health check server:", err)
		}
	}()

	eventQueue := make(chan api.Record, 10)
	fallBackQueue := make(chan api.Record, 100)

	var wg sync.WaitGroup
	numWorkers := 5

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go api.Worker(i, eventQueue, &wg)
	}

	go api.ProcessFallbackQueue(eventQueue, fallBackQueue)

	start := time.Now()
	// FIGURE OUT HOW TO SEND THIS BACK FROM WORKER POOL
	recordCount := 0

	go func() {
		ctx := context.Background()
		if err := api.StartFirehose(ctx, eventQueue, fallBackQueue, server, handle, apikey); err != nil {
			slog.Error("Firehose encountered an error", "error", err)
		}
		close(eventQueue)
	}()

	wg.Wait()

	close(fallBackQueue)

	recordRate := recordCount / int(time.Since(start))
	slog.Info("Average records per second", "", recordRate)

	slog.Info("Application stopped")
}
