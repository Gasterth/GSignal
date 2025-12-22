package gsignal

import "sync"

// subscription is the concrete implementation of the Subscription interface.
// It is a push-only channel managed internally by the dispatcher.
type subscription struct {
	mu     sync.RWMutex
	events chan Event
	closed bool
}

//──────────────────────────────────────────────
// Constructor
//──────────────────────────────────────────────

func newSubscription() *subscription {
	return &subscription{
		events: make(chan Event, 64), // default buffer
	}
}

//──────────────────────────────────────────────
// Subscription interface methods
//──────────────────────────────────────────────

// Events returns the read-only channel where events arrive.
func (s *subscription) Events() <-chan Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.events
}

// Close closes the subscription.
// It is idempotent and does not panic.
func (s *subscription) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true
	close(s.events)
	return nil
}

// IsClosed reports whether the subscription is closed.
func (s *subscription) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

//──────────────────────────────────────────────
// Internal API for Dispatcher
//──────────────────────────────────────────────

// push tries to send an event to the channel.
// If the buffer is full, the event is silently dropped.
func (s *subscription) push(evt Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return
	}

	select {
	case s.events <- evt:
	default:
		// silent drop (explicit policy)
	}
}
