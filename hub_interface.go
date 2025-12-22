package gsignal

import "context"

//──────────────────────────────────────────────
// Subscription Interface
//──────────────────────────────────────────────

// Subscription represents a connection to an event stream.
// Events are received via a read-only channel.
type Subscription interface {
	// Events returns the channel to read events from.
	Events() <-chan Event

	// Close closes the subscription.
	// It is idempotent.
	Close() error

	// IsClosed reports whether the subscription is closed.
	IsClosed() bool
}

//──────────────────────────────────────────────
// Hub Interface (public)
//──────────────────────────────────────────────

// Hub defines the public interface of GSignal.
// It is thread-safe and designed for maximum testability.
type Hub interface {
	// Publish sends an event synchronously.
	// Returns ErrHubClosed if the hub is closed.
	Publish(evt Event) error

	// PublishAsync inserts the event into the async queue.
	// Returns ErrHubClosed if the hub is closed.
	PublishAsync(evt Event) error

	// Subscribe creates a new subscription.
	// If no types are specified -> global subscription.
	Subscribe(ctx context.Context, types ...string) (Subscription, error)

	// Unsubscribe forces the removal of a subscription.
	Unsubscribe(sub Subscription) error

	// Shutdown closes the hub and all subscriptions.
	// It is idempotent.
	Shutdown() error

	// Closed indicates whether the hub has been closed.
	Closed() bool
}
