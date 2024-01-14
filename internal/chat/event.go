package chat

import (
	"context"
	"errors"
	"fmt"

	"github.com/Salam4nder/chat/internal/event"
	"github.com/google/uuid"
)

const (
	MessageCreatedInRoomEventName = "MessageCreatedInRoom"
)

var ErrInvalidEventError = errors.New("invalid event")

type MessageCreatedInRoomEvent struct {
	eventID uuid.UUID
	payload event.Payload
}

// ID returns the event ID.
func (x MessageCreatedInRoomEvent) ID() uuid.UUID {
	return x.eventID
}

// Payload returns the event payload.
func (x MessageCreatedInRoomEvent) Payload() event.Payload {
	return x.payload
}

// Name returns the name of the event that the handlers will listen to.
func (x MessageCreatedInRoomEvent) EventName() string {
	return MessageCreatedInRoomEventName
}

func HandleMessageCreatedInRoomEvent(ctx context.Context, event event.Event) error {
	payload, ok := event.Payload().(Message)
	if !ok {
		return ErrInvalidEventError
	}

	if err := payload.Valid(); err != nil {
		return fmt.Errorf("validating message: %w", err)
	}

	return nil
}
