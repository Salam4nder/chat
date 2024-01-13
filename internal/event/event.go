package event

import (
	"time"

	"github.com/google/uuid"
)

// Payload defines an event payload.
type Payload any

// IEvent defines an abstract event.
type IEvent interface {
	ID() string
	// Name returns the event name
	// that the handlers will subscribe to.
	Name() string
	Payload() Payload
}

// Event defines an event.
type Event struct {
	id        string
	name      string
	payload   Payload
	occuredAt time.Time
}

// New returns a new event.
func New(name string, payload Payload) Event {
	return Event{
		id:        uuid.NewString(),
		name:      name,
		payload:   payload,
		occuredAt: time.Now(),
	}
}

// ID returns the event ID.
func (x Event) ID() string {
	return x.id
}

// Name returns the event name.
func (x Event) Name() string {
	return x.name
}

// Payload returns the event payload.
func (x Event) Payload() Payload {
	return x.payload
}
