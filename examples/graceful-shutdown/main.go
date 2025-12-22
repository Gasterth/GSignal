package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Merluz/gsignal"
)

func main() {
	hub := gsignal.New()

	ctx := context.Background()

	sub, err := hub.Subscribe(ctx, "system.event")
	if err != nil {
		panic(err)
	}

	// Consumer
	go func() {
		for evt := range sub.Events() {
			fmt.Println("received:", evt.Payload)
			time.Sleep(100 * time.Millisecond)
		}
		fmt.Println("subscription closed")
	}()

	// Publish events
	for i := 0; i < 5; i++ {
		evt, _ := gsignal.NewEvent("system.event", i)
		_ = hub.PublishAsync(evt)
	}

	// Simulate application runtime
	time.Sleep(500 * time.Millisecond)

	fmt.Println("shutting down hub...")
	_ = hub.Shutdown()

	// Allow time for the consumer to exit cleanly
	time.Sleep(300 * time.Millisecond)
	fmt.Println("shutdown completed")
}
