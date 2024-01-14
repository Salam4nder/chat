package chat

import (
	"context"
	"errors"
	"fmt"

	db "github.com/Salam4nder/chat/internal/db/keyspace/chat"
	"github.com/Salam4nder/chat/internal/event"
	"github.com/google/uuid"
)

const (
	MessageCreatedInRoomEventName = "MessageCreatedInRoom"
)

var (
	ErrWrongEventType    = errors.New("wrong event type")
	ErrInvalidEventError = errors.New("invalid event")
)

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

func (x *MessageService) HandleMessageCreatedInRoomEvent(
	ctx context.Context,
	event event.Event,
) error {
	payload, ok := event.Payload().(Message)
	if !ok {
		return ErrWrongEventType
	}

	if err := payload.Valid(); err != nil {
		return fmt.Errorf("chat: %w: %w", ErrInvalidEventError, err)
	}

	if err := x.messageRepo.CreateMessageByRoom(ctx, db.CreateMessageByRoomParams{
		Data:   payload.Body,
		Type:   payload.TypeString(),
		Sender: payload.Author,
		RoomID: payload.RoomID,
	}); err != nil {
		return fmt.Errorf("message service: persisting message in room, %w", err)
	}

	return nil
}
