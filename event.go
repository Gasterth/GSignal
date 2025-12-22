package gsignal

import (
	"time"

	"github.com/oklog/ulid/v2"
)

//──────────────────────────────────────────────
// Event
//──────────────────────────────────────────────

// Event represents an immutable message carried by the hub.
type Event struct {
	ID        string
	Type      string
	Timestamp time.Time
	Payload   any
	Metadata  map[string]string
}

//──────────────────────────────────────────────
// Constructors
//──────────────────────────────────────────────

// NewEvent creates a new event with:
// - ULID ID
// - UTC timestamp
// - optional payload
func NewEvent(eventType string, payload any) (Event, error) {
	if eventType == "" {
		return Event{}, ErrInvalidEvent
	}

	return Event{
		ID:        ulid.Make().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}, nil
}

// WithMetadata returns a copy of the event with additional metadata.
func (e Event) WithMetadata(key, value string) Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}
