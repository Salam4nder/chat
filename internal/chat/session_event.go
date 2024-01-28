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
	ErrUserIDInvalid   = errors.New("user id invalid")
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
func NewSessionService(
	natsClient *nats.Conn,
	registry *event.Registry,
) *SessionService {
	return &SessionService{
		natsClient: natsClient,
		registry:   registry,
	}
}

// SessionConnectedPayload is the payload for
// a SessionConnectedEvent.
type SessionConnectedPayload struct {
	UserID   string
	RoomID   string
	Username string
	Conn     *websocket.Conn
}

// Valid returns nil if the payload is valid.
func (x SessionConnectedPayload) Valid() error {
	var userIDErr, roomIDErr, usernameErr, connErr error

	if x.UserID == "" {
		userIDErr = ErrUserIDInvalid
	}
	if x.RoomID == "" {
		roomIDErr = ErrRoomIDInvalid
	}
	if x.Username == "" {
		usernameErr = ErrUsernameInvalid
	}
	if x.Conn == nil {
		connErr = ErrConnInvalid
	}

	return errors.Join(userIDErr, roomIDErr, usernameErr, connErr)
}

// HandleSessionConnectedEvent handles a new session connected event.
func (x *SessionService) HandleSessionConnectedEvent(
	ctx context.Context,
	evt event.Event,
) error {
	log.Info().Msg("HandleNewSessionConnectedEvent ->")
	defer log.Info().Msg("HandleNewSessionConnectedEvent <-")

	payload, ok := evt.Payload.(SessionConnectedPayload)
	if !ok {
		return event.ErrInvalidEventType
	}

	if err := payload.Valid(); err != nil {
		return fmt.Errorf(
			"chat: %w, %w",
			event.ErrInvalidEventPayloadError,
			err,
		)
	}

	room, exists := ChatRomoms[payload.RoomID]
	if !exists {
		room = NewRoom(&payload.RoomID, x.registry)
		ChatRomoms[payload.RoomID] = room
		go room.Run(ctx)
	}

	session := &UserSess{
		UserID:      payload.UserID,
		RoomID:      payload.RoomID,
		DisplayName: payload.Username,
		Conn:        payload.Conn,
	}

	room.Join <- session

	return nil
}
