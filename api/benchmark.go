package api

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

type Event struct {
	ID int
}

func RunBenchmark() {
	rand.Seed(uint64(time.Now().UnixNano()))

	numEvents := 10000

	primaryBufferSizes := []int{100, 1000}
	fallbackBufferSizes := []int{100, 1000, 5000}
	numWorkersList := []int{1, 2, 4, 8, 16}

	// primaryBufferSizes := []int{100}
	// fallbackBufferSizes := []int{100}
	// numWorkersList := []int{4}

	slog.Info("Starting benchmark...")

	for _, primaryBufferSize := range primaryBufferSizes {
		for _, fallbackBufferSize := range fallbackBufferSizes {
			for _, numWorkers := range numWorkersList {
				duration, dropped := simulate(primaryBufferSize, fallbackBufferSize, numWorkers, numEvents)
				slog.Info("Results:", "Primary buffer size", primaryBufferSize, "Fallback buffer size:", fallbackBufferSize, "Threads", numWorkers, "Duration", duration, "Dropped events", dropped)
			}
		}

	}
}

func simulate(primaryBufferSizeufferSize, fallbackBufferSize, numWorkers, numEvents int) (time.Duration, int) {
	primary := make(chan Record, primaryBufferSizeufferSize)
	fallback := make(chan Record, fallbackBufferSize)
	droppedEvents := 0

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		go Worker(i, primary, &wg)
	}
	go ProcessFallbackQueue(primary, fallback)

	start := time.Now()

	// go func() {
	// 	ctx := context.Background()
	// 	if err := api.StartFirehose(ctx, eventQueue, fallBackQueue, server, handle, apikey); err != nil {
	// 		slog.Error("Firehose encountered an error", "error", err)
	// 	}
	// 	close(primary)
	// }()

	for i := 0; i < numEvents; i++ {
		event := Record{ctx: context.Background()}

		// select {
		// case primary <- event:
		// default:
		// 	fallback <- event
		// }
		select {
		case primary <- event:
			// slog.Info("Queue lengths", "Primary queue", len(primary), "Fallback queue", len(fallback))
		default:
			// slog.Warn("Primary queue full. Record sent to fallback queue", "", record)
			select {
			case fallback <- event:
				// slog.Info("Queue lengths", "Primary queue", len(eventQueue), "Fallback queue", len(fallBackQueue))
				// slog.Info("Record enqueued to fallback queue", "", record)
			default:
				// slog.Error("Fallback queue also full. Dropping record", "", event)
				droppedEvents += 1
			}
		}
	}

	wg.Wait()

	close(fallback)

	return time.Since(start), droppedEvents
}
