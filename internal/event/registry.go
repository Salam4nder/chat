package event

import (
	"context"
	"sync"
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
func (x *Registry) Subscribe(event Event, handler Handler) {
	x.mu.Lock()
	defer x.mu.Unlock()

	x.handlers[event.Name()] = append(x.handlers[event.Name()], handler)
}

// Publish publishes the given event to all its handlers.
func (x *Registry) Publish(ctx context.Context, event Event) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	for _, handler := range x.handlers[event.Name()] {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}

	return nil
}
