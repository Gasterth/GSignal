package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Merluz/gsignal"
)

func main() {
	hub := gsignal.New()
	defer hub.Shutdown()

	ctx := context.Background()

	sub, err := hub.Subscribe(ctx, "job.done")
	if err != nil {
		panic(err)
	}

	// Consumer
	go func() {
		for evt := range sub.Events() {
			fmt.Println("received:", evt.Payload)
			// Simulate slow work
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// ──────────────────────────────────────────
	// Publish SYNC
	// ──────────────────────────────────────────
	fmt.Println("publish sync start")
	start := time.Now()

	for i := 0; i < 3; i++ {
		evt, _ := gsignal.NewEvent("job.done", fmt.Sprintf("sync-%d", i))
		_ = hub.Publish(evt)
	}

	fmt.Println("publish sync end   (elapsed:", time.Since(start), ")")

	// ──────────────────────────────────────────
	// Publish ASYNC
	// ──────────────────────────────────────────
	fmt.Println("publish async start")
	start = time.Now()

	for i := 0; i < 3; i++ {
		evt, _ := gsignal.NewEvent("job.done", fmt.Sprintf("async-%d", i))
		_ = hub.PublishAsync(evt)
	}

	fmt.Println("publish async end  (elapsed:", time.Since(start), ")")

	// Give the consumer time to finish
	time.Sleep(1 * time.Second)
}
