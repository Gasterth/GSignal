package gsignal

import (
	"sync"
	"time"
)

//──────────────────────────────────────────────
// AsyncDispatcher -> queue + worker pool
//──────────────────────────────────────────────

type AsyncDispatcher struct {
	queue   chan Event
	workers int

	dispatcher *Dispatcher

	stopOnce sync.Once
	stopped  chan struct{}
}

//──────────────────────────────────────────────
// Constructor
//──────────────────────────────────────────────

// NewAsyncDispatcher creates an async dispatcher with a worker pool.
// Uses sensible default values (queue + workers).
func NewAsyncDispatcher(dispatcher *Dispatcher) *AsyncDispatcher {
	ad := &AsyncDispatcher{
		queue:      make(chan Event, 256), // default buffer
		workers:    4,                     // default workers
		dispatcher: dispatcher,
		stopped:    make(chan struct{}),
	}

	ad.startWorkers()
	return ad
}

//──────────────────────────────────────────────
// Public APIs
//──────────────────────────────────────────────

// Enqueue inserts an event into the async queue.
// If the queue is full, the event is silently dropped.
func (ad *AsyncDispatcher) Enqueue(evt Event) {
	select {
	case ad.queue <- evt:
		return
	default:
		// silent drop (explicit policy)
		return
	}
}

// Stop closes the queue and waits for all workers to finish.
// It is idempotent.
func (ad *AsyncDispatcher) Stop() {
	ad.stopOnce.Do(func() {
		close(ad.queue)
		<-ad.stopped
	})
}

//──────────────────────────────────────────────
// Worker logic
//──────────────────────────────────────────────

func (ad *AsyncDispatcher) startWorkers() {
	var wg sync.WaitGroup
	wg.Add(ad.workers)

	for i := 0; i < ad.workers; i++ {
		go func() {
			defer wg.Done()

			for evt := range ad.queue {
				_ = ad.dispatcher.Dispatch(evt)

				// micro sleep to avoid extreme bursts
				time.Sleep(50 * time.Microsecond)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(ad.stopped)
	}()
}
