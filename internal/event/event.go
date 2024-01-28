package event

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidEventType         = errors.New("event type invalid")
	ErrInvalidEventPayloadError = errors.New("event payload invalid")
)

// Handler defines an event handler.
// Errors are logged and ignored.
type Handler func(ctx context.Context, evt Event)

// Payload defines an event payload.
type Payload any

// Event defines an event.
type Event struct {
	ID        string
	Name      string
	Payload   Payload
	OccuredAt time.Time
}

// New returns a new event.
func New(name string, payload Payload) Event {
	return Event{
		ID:        uuid.NewString(),
		Name:      name,
		Payload:   payload,
		OccuredAt: time.Now(),
	}
}
