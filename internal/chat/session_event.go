package chat

import (
	"context"
	"errors"
	"fmt"

	"github.com/Salam4nder/chat/internal/event"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

const SessionConnectedEvent = "session_connected"

var (
	ErrRoomIDInvalid   = errors.New("room id invalid")
	ErrUsernameInvalid = errors.New("username invalid")
	ErrConnInvalid     = errors.New("conn invalid")
)

// SessionService handles session events and communicates with NATS.
type SessionService struct {
	natsClient *nats.Conn
	registry   *event.Registry
}

// NewSessionService creates a new SessionService.
// It handles session events and communicates with NATS.
func NewSessionService(natsClient *nats.Conn, registry *event.Registry) *SessionService {
	return &SessionService{natsClient: natsClient, registry: registry}
}

type SessionConnectedPayload struct {
	RoomID   string
	Username string
	Conn     *websocket.Conn
}

func (x SessionConnectedPayload) Valid() error {
	if x.RoomID == "" {
		return ErrRoomIDInvalid
	}
	if x.Username == "" {
		return ErrUsernameInvalid
	}
	if x.Conn == nil {
		return ErrConnInvalid
	}
	return nil
}

// HandleSessionConnectedEvent handles a new session connected event.
func (x *SessionService) HandleSessionConnectedEvent(ctx context.Context, evt event.Event) error {
	log.Info().Msg("HandleNewSessionConnectedEvent ->")
	defer log.Info().Msg("HandleNewSessionConnectedEvent <-")

	payload, ok := evt.Payload.(SessionConnectedPayload)
	if !ok {
		return event.ErrInvalidEventType
	}

	if err := payload.Valid(); err != nil {
		return fmt.Errorf("chat: %w, %w", event.ErrInvalidEventPayloadError, err)
	}

	room, exists := Rooms[payload.RoomID]
	if !exists {
		room = NewRoom(&payload.RoomID)
		Rooms[payload.RoomID] = room
		go room.Run()
	}

	session := NewSession(
		Rooms[payload.RoomID],
		payload.Conn,
		payload.Username,
		x.registry,
	)
	room.Join <- session
	go session.Read(ctx)
	go session.Write()

	return nil
}
