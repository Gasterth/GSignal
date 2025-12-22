package gsignal

import (
	"context"
	"sync"
)

//──────────────────────────────────────────────
// Hub Implementation
//──────────────────────────────────────────────

type hub struct {
	mu sync.RWMutex

	dispatcher   *Dispatcher
	asyncWorkers *AsyncDispatcher

	isClosed bool
}

//──────────────────────────────────────────────
// Constructor
//──────────────────────────────────────────────

// New creates a new Event Hub.
// The hub is thread-safe and ready to use.
func New() Hub {
	dispatcher := NewDispatcher()

	return &hub{
		dispatcher:   dispatcher,
		asyncWorkers: NewAsyncDispatcher(dispatcher),
		isClosed:     false,
	}
}

//──────────────────────────────────────────────
// Publish (sync)
//──────────────────────────────────────────────

// Publish sends an event synchronously.
// Returns ErrHubClosed if the hub is closed.
func (h *hub) Publish(evt Event) error {
	h.mu.RLock()
	closed := h.isClosed
	dispatcher := h.dispatcher
	h.mu.RUnlock()

	if closed {
		return ErrHubClosed
	}

	return dispatcher.Dispatch(evt)
}

//──────────────────────────────────────────────
// Publish (async)
//──────────────────────────────────────────────

// PublishAsync inserts the event into the async queue.
// Returns ErrHubClosed if the hub is closed.
func (h *hub) PublishAsync(evt Event) error {
	h.mu.RLock()
	closed := h.isClosed
	workers := h.asyncWorkers
	h.mu.RUnlock()

	if closed {
		return ErrHubClosed
	}

	workers.Enqueue(evt)
	return nil
}

//──────────────────────────────────────────────
// Subscribe
//──────────────────────────────────────────────

// Subscribe creates a new subscription.
// If no types are provided -> global subscription.
func (h *hub) Subscribe(ctx context.Context, types ...string) (Subscription, error) {
	h.mu.RLock()
	if h.isClosed {
		h.mu.RUnlock()
		return nil, ErrHubClosed
	}
	dispatcher := h.dispatcher
	h.mu.RUnlock()

	sub := newSubscription()

	if len(types) == 0 {
		dispatcher.AddAll(sub)
	} else {
		for _, t := range types {
			dispatcher.Add(t, sub)
		}
	}

	// Auto-close on context cancellation
	go func() {
		<-ctx.Done()
		_ = sub.Close()
	}()

	return sub, nil
}

//──────────────────────────────────────────────
// Unsubscribe
//──────────────────────────────────────────────

// Unsubscribe forces the removal of a subscription.
func (h *hub) Unsubscribe(sub Subscription) error {
	impl, ok := sub.(*subscription)
	if !ok {
		return ErrInvalidSubscribe
	}

	h.mu.RLock()
	closed := h.isClosed
	dispatcher := h.dispatcher
	h.mu.RUnlock()

	if closed {
		return ErrHubClosed
	}

	dispatcher.RemoveAll(impl)
	_ = impl.Close()

	return nil
}

//──────────────────────────────────────────────
// Shutdown
//──────────────────────────────────────────────

// Shutdown closes the hub and all subscriptions.
// It is idempotent.
func (h *hub) Shutdown() error {
	h.mu.Lock()
	if h.isClosed {
		h.mu.Unlock()
		return nil
	}
	h.isClosed = true

	dispatcher := h.dispatcher
	workers := h.asyncWorkers
	h.mu.Unlock()

	workers.Stop()

	for _, s := range dispatcher.AllSubscriptions() {
		_ = s.Close()
	}

	return nil
}

// Closed indicates whether the hub has been closed.
func (h *hub) Closed() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isClosed
}
