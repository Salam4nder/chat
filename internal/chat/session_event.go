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

const NewSessionConnectedEvent = "new_session_connected"

var (
	ErrRoomIDInvalid   = errors.New("room id invalid")
	ErrUsernameInvalid = errors.New("username invalid")
	ErrConnInvalid     = errors.New("conn invalid")
)

// SessionService handles session events and communicates with NATS.
type SessionService struct {
	natsClient *nats.Conn
}

// NewSessionService creates a new SessionService.
// It handles session events and communicates with NATS.
func NewSessionService(natsClient *nats.Conn) *SessionService {
	return &SessionService{natsClient: natsClient}
}

type NewSessionConnectedEventPayload struct {
	RoomID   string
	Username string
	Conn     *websocket.Conn
}

func (x NewSessionConnectedEventPayload) Valid() error {
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

// HandleNewSessionConnectedEvent handles a new session connected event.
func (s *SessionService) HandleNewSessionConnectedEvent(ctx context.Context, evt event.Event) error {
	log.Info().Msg("HandleNewSessionConnectedEvent ->")
	defer log.Info().Msg("HandleNewSessionConnectedEvent <-")

	payload, ok := evt.Payload.(NewSessionConnectedEventPayload)
	if !ok {
		return event.ErrWrongEventType
	}

	if err := payload.Valid(); err != nil {
		return fmt.Errorf("chat: %w, %w", event.ErrInvalidEventError, err)
	}

	if room, exists := Rooms[payload.RoomID]; exists {
		session := NewSession(
			Rooms[payload.RoomID],
			payload.Conn,
			payload.Username,
		)
		room.Join <- session
		go session.Read()
		go session.Write()
		return nil
	}

	room := NewRoom()
	Rooms[payload.RoomID] = room
	go room.Run()

	session := NewSession(
		Rooms[payload.RoomID],
		payload.Conn,
		payload.Username,
	)
	go session.Read()
	go session.Write()
	room.Join <- session

	return nil
}
