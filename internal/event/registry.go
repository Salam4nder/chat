package event

import (
	"sync"

	"github.com/rs/zerolog/log"
)

// Registry is a concurrent-safe registry that holds events and their handlers.
// It is used for in-memory pub/sub. Use cases include domain side effects and
// event sourcing.
type Registry struct {
	mu       sync.Mutex
	handlers map[string][]Handler
}

// NewRegistry returns a new Registry, ready to be used.
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe adds a new handler for the given event.
func (x *Registry) Subscribe(eventName string, handler Handler) {
	x.mu.Lock()
	defer x.mu.Unlock()

	x.handlers[eventName] = append(x.handlers[eventName], handler)
}

// Publish publishes the given event to all its handlers.
func (x *Registry) Publish(event Event) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	for _, handler := range x.handlers[event.Name] {
		if err := handler(event); err != nil {
			log.Error().Err(err).Msg("event: handling event")
			return err
		}
	}

	return nil
}
