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

	// Cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Subscription bound to the context
	sub, err := hub.Subscribe(ctx, "task.finished")
	if err != nil {
		panic(err)
	}

	// Consumer
	go func() {
		for evt := range sub.Events() {
			fmt.Println("received:", evt.Payload)
		}
		fmt.Println("subscription closed")
	}()

	// Publish event
	evt, _ := gsignal.NewEvent("task.finished", "first event")
	_ = hub.Publish(evt)

	time.Sleep(100 * time.Millisecond)

	// Cancel the context -> the subscription closes automatically
	fmt.Println("cancelling context")
	cancel()

	time.Sleep(100 * time.Millisecond)

	// This event will NOT be received
	evt2, _ := gsignal.NewEvent("task.finished", "second event")
	_ = hub.Publish(evt2)

	time.Sleep(200 * time.Millisecond)
}
