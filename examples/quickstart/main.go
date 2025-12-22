package main

import (
	"context"
	"fmt"

	"github.com/Merluz/gsignal"
)

func main() {
	// Create the hub
	hub := gsignal.New()
	defer hub.Shutdown()

	// Application context
	ctx := context.Background()

	// Subscription to an event type
	sub, err := hub.Subscribe(ctx, "user.created")
	if err != nil {
		panic(err)
	}

	// Create an event
	evt, err := gsignal.NewEvent("user.created", map[string]string{
		"id":   "42",
		"name": "Merluz",
	})
	if err != nil {
		panic(err)
	}

	// Synchronous publication
	if err := hub.Publish(evt); err != nil {
		panic(err)
	}

	// Receive event
	received := <-sub.Events()
	fmt.Println("received event:")
	fmt.Println("  type:", received.Type)
	fmt.Println("  payload:", received.Payload)
}
