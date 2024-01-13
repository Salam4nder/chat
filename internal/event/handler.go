package event

import (
	"context"
)

// Handler defines an event handler.
type Handler func(ctx context.Context, event Event) error
