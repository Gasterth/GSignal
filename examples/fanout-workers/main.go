package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Merluz/gsignal"
)

func main() {
	hub := gsignal.New()
	defer hub.Shutdown()

	ctx := context.Background()

	// Shared subscription
	sub, err := hub.Subscribe(ctx, "job.process")
	if err != nil {
		panic(err)
	}

	const workers = 3
	var wg sync.WaitGroup
	wg.Add(workers)

	// Worker pool consuming from the same event stream
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			for evt := range sub.Events() {
				fmt.Printf("worker %d processing: %v\n", id, evt.Payload)
				time.Sleep(150 * time.Millisecond)
			}
		}(i)
	}

	// Publish events
	for i := 0; i < 6; i++ {
		evt, _ := gsignal.NewEvent("job.process", i)
		_ = hub.PublishAsync(evt)
	}

	// Give workers time to process
	time.Sleep(2 * time.Second)

	// Close everything
	_ = hub.Unsubscribe(sub)
	wg.Wait()
}
