package chat

import (
	"errors"
	"time"

	"github.com/Salam4nder/chat/internal/db/keyspace/chat"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Session is a single client session in a room.
type Session struct {
	ID     string
	Active bool
	Room   *Room
	Conn   *websocket.Conn
	In     chan Message
	// FriendlyName is the displayed name of the connected user.
	FriendlyName string
	UserID       uuid.UUID

	MessageRepo chat.MessageRepository
}

// NewSession returns a new session.
func NewSession(
	room *Room,
	conn *websocket.Conn,
	username string,
) *Session {
	return &Session{
		ID:           uuid.NewString(),
		Active:       true,
		Room:         room,
		Conn:         conn,
		In:           make(chan Message),
		FriendlyName: username,
	}
}

// Write writes messages to the websocket connection.
func (x *Session) Write() {
	x.writePump()
}

// Read reads messages from the websocket connection.
func (x *Session) Read() {
	x.readPump()
}

func (x *Session) readPump() {
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

		x.Room.Broadcast <- Message{
			Type:      mType,
			RoomID:    x.Room.ID,
			SessionID: x.ID,
			Body:      m,
			Author:    x.FriendlyName,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
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
