package chat

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	db "github.com/Salam4nder/chat/internal/db/keyspace/chat"
	"github.com/Salam4nder/chat/internal/event"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

const (
	MessageCreatedInRoomEvent = "MessageCreatedInRoom"
)

// MessageService defines the main message service.
// It can persist messages and communicates with NATS.
type MessageService struct {
	messageRepo db.MessageRepository
	natsClient  *nats.Conn
}

// NewMessageService returns a new instance of MessageService.
// It can persist messages and communicate with NATS.
func NewMessageService(repo db.MessageRepository, client *nats.Conn) *MessageService {
	return &MessageService{
		messageRepo: repo,
		natsClient:  client,
	}
}

func (x *MessageService) HandleMessageCreatedInRoomEvent(
	evt event.Event,
) error {
	log.Info().Msg("HandleMessageCreatedInRoomEvent ->")
	defer log.Info().Msg("<- HandleMessageCreatedInRoomEvent")

	payload, ok := evt.Payload.(Message)
	if !ok {
		return event.ErrInvalidEventType
	}

	if err := payload.Valid(); err != nil {
		return fmt.Errorf("chat: %w: %w", event.ErrInvalidEventPayloadError, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := x.messageRepo.CreateMessageByRoom(ctx, db.CreateMessageByRoomParams{
		Data:   payload.Body,
		Type:   payload.TypeString(),
		Sender: payload.Author,
		RoomID: payload.RoomID,
	}); err != nil {
		return fmt.Errorf("message service: persisting message in room, %w", err)
	}

	buf := bytes.Buffer{}
	if err := gob.NewEncoder(&buf).Encode(payload); err != nil {
		return fmt.Errorf("message service: encoding event, %w", err)
	}

	if err := x.natsClient.Publish(evt.Name, buf.Bytes()); err != nil {
		return fmt.Errorf("message service: publishing event, %w", err)
	}

	return nil
}
