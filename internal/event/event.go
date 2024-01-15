package event

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWrongEventType    = errors.New("wrong event type")
	ErrInvalidEventError = errors.New("invalid event")
)

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
