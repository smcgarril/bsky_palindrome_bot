package api

import (
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

type Event struct {
	ID int
}

func RunBenchmark() {
	rand.Seed(time.Now().UnixNano())

	numEvents := 1000

	primaryBufferSizes := []int{100, 1000}
	fallbackBufferSizes := []int{100, 1000, 5000}
	numWorkersList := []int{1, 2, 4, 8, 16}

	slog.Info("Starting benchmark...")

	for _, primaryBufferSize := range primaryBufferSizes {
		for _, fallbackBufferSize := range fallbackBufferSizes {
			for _, numWorkers := range numWorkersList {
				duration, dropped := simulate(primaryBufferSize, fallbackBufferSize, numWorkers, numEvents)
				slog.Info("Results:",
					"Primary buffer size", primaryBufferSize,
					"Fallback buffer size", fallbackBufferSize,
					"Threads", numWorkers,
					"Duration", duration,
					"Dropped events", dropped)
			}
		}
	}
}

func simulate(primaryBufferSize, fallbackBufferSize, numWorkers, numEvents int) (time.Duration, int) {
	primary := make(chan Event, primaryBufferSize)
	fallback := make(chan Event, fallbackBufferSize)
	droppedEvents := 0

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, primary, &wg)
	}

	// Process fallback queue
	go func() {
		for event := range fallback {
			processEvent(event)
		}
	}()

	start := time.Now()

	// Simulate event generation with realistic traffic pattern
	go func() {
		for i := 0; i < numEvents; i++ {
			event := Event{ID: i}
			// Simulate average 50 events/sec with spikes
			time.Sleep(eventInterval())

			select {
			case primary <- event:
			default:
				select {
				case fallback <- event:
				default:
					droppedEvents++
				}
			}
		}
		close(primary)
	}()

	// Wait for all workers to finish
	wg.Wait()
	close(fallback)

	return time.Since(start), droppedEvents
}

func eventInterval() time.Duration {
	// Average 50 events/sec with spikes up to 100+/sec
	baseInterval := 20 * time.Millisecond // 50 events/sec
	if rand.Float64() < 0.1 {             // 10% chance of spike
		return time.Duration(rand.Intn(10)+1) * time.Millisecond // Spike: 100-200 events/sec
	}
	return baseInterval
}

func worker(id int, events <-chan Event, wg *sync.WaitGroup) {
	defer wg.Done()
	for event := range events {
		processEvent(event)
	}
}

func processEvent(event Event) {
	// Simulate network call with mostly fast responses but occasional delays
	delay := time.Millisecond * time.Duration(rand.Intn(20)+5) // 5-25 ms typical
	if rand.Float64() < 0.05 {                                 // 5% chance of significant delay
		delay = time.Millisecond * time.Duration(rand.Intn(500)+200) // 200-700 ms
	}
	time.Sleep(delay)
}
