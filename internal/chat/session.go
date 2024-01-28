package chat

import (
	"context"
	"errors"
	"time"

	"github.com/Salam4nder/chat/internal/event"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Session is a single client session in a chat room.
type Session struct {
	ID     string
	Active bool
	Room   *Room
	Conn   *websocket.Conn
	In     chan Message
	// Username is the displayed name of the connected user.
	Username string
	UserID   uuid.UUID

	registry *event.Registry
}

// NewSession returns a new session.
func NewSession(
	room *Room,
	conn *websocket.Conn,
	username string,
	registry *event.Registry,
) *Session {
	return &Session{
		ID:       uuid.NewString(),
		Active:   true,
		Room:     room,
		Conn:     conn,
		In:       make(chan Message),
		Username: username,
		registry: registry,
	}
}

// Write writes messages to the websocket connection.
func (x *Session) Write() {
	x.writePump()
}

// Read reads messages from the websocket connection.
func (x *Session) Read(ctx context.Context) {
	x.readPump(ctx)
}

func (x *Session) readPump(ctx context.Context) {
	defer func() {
		x.Active = false
		x.Room.Leave <- x
	}()

	for {
		mType, m, err := x.Conn.ReadMessage()
		if err != nil {
			var closeErr *websocket.CloseError
			if errors.As(err, &closeErr) {
				log.Info().
					Int("code", closeErr.Code).
					Str("text", closeErr.Text).
					Msg("chat: close message received")
				break
			}
		}

		message := Message{
			ID:        uuid.New(),
			Type:      mType,
			RoomID:    x.Room.ID,
			SessionID: x.ID,
			Body:      m,
			Author:    x.Username,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		if err := x.registry.Publish(
			ctx,
			event.New(MessageCreatedInRoomEvent, message),
		); err != nil {
			log.Error().Err(err).Msg("chat: publishing message")
		}

		x.Room.Broadcast <- message
	}
}

func (x *Session) writePump() {
	defer x.Conn.Close()

	for message := range x.In {
		err := x.Conn.WriteMessage(message.Type, message.Body)
		if err != nil {
			log.Error().Err(err).Msg("chat: writing message")
			continue
		}
	}
}
