package gsignal

import "sync"

// Dispatcher is responsible for delivering events
// to all interested subscriptions.
type Dispatcher struct {
	mu sync.RWMutex

	// Subscribers by single event type
	byType map[string][]*subscription

	// Global subscribers -> receive ALL events
	all []*subscription
}

//──────────────────────────────────────────────
// Constructor
//──────────────────────────────────────────────

// NewDispatcher creates a new empty dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		byType: make(map[string][]*subscription),
		all:    make([]*subscription, 0),
	}
}

//──────────────────────────────────────────────
// Subscription registration
//──────────────────────────────────────────────

// Add registers a subscription for a specific event type.
func (d *Dispatcher) Add(eventType string, sub *subscription) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.byType[eventType] = append(d.byType[eventType], sub)
}

// AddAll registers a global subscription (receives all events).
func (d *Dispatcher) AddAll(sub *subscription) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.all = append(d.all, sub)
}

//──────────────────────────────────────────────
// Subscription removal
//──────────────────────────────────────────────

// Remove removes a subscription from a specific event type.
func (d *Dispatcher) Remove(eventType string, sub *subscription) {
	d.mu.Lock()
	defer d.mu.Unlock()

	list := d.byType[eventType]
	out := list[:0]

	for _, s := range list {
		if s != sub {
			out = append(out, s)
		}
	}

	if len(out) == 0 {
		delete(d.byType, eventType)
	} else {
		d.byType[eventType] = out
	}
}

// RemoveAll removes a subscription from all lists.
func (d *Dispatcher) RemoveAll(sub *subscription) {
	d.mu.Lock()
	defer d.mu.Unlock()

	out := d.all[:0]
	for _, s := range d.all {
		if s != sub {
			out = append(out, s)
		}
	}
	d.all = out

	for key, list := range d.byType {
		out = list[:0]
		for _, s := range list {
			if s != sub {
				out = append(out, s)
			}
		}

		if len(out) == 0 {
			delete(d.byType, key)
		} else {
			d.byType[key] = out
		}
	}
}

//──────────────────────────────────────────────
// Introspection
//──────────────────────────────────────────────

// AllSubscriptions returns a read-only snapshot
// of ALL registered subscriptions.
func (d *Dispatcher) AllSubscriptions() []*subscription {
	d.mu.RLock()
	defer d.mu.RUnlock()

	out := make([]*subscription, 0)
	seen := make(map[*subscription]struct{})

	for _, sub := range d.all {
		if _, ok := seen[sub]; !ok {
			seen[sub] = struct{}{}
			out = append(out, sub)
		}
	}

	for _, list := range d.byType {
		for _, sub := range list {
			if _, ok := seen[sub]; !ok {
				seen[sub] = struct{}{}
				out = append(out, sub)
			}
		}
	}

	return out
}

//──────────────────────────────────────────────
// Dispatch
//──────────────────────────────────────────────

// Dispatch sends an event to all relevant subscribers.
// It never blocks and never returns errors.
func (d *Dispatcher) Dispatch(evt Event) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Global subscribers
	for _, sub := range d.all {
		if !sub.IsClosed() {
			sub.push(evt)
		}
	}

	// Specific subscribers
	for _, sub := range d.byType[evt.Type] {
		if !sub.IsClosed() {
			sub.push(evt)
		}
	}

	return nil
}
