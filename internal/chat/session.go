package chat

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Session is a single client session in a room.
type Session struct {
	ID     uuid.UUID
	Active bool
	Room   *Room
	Conn   *websocket.Conn
	In     chan Message
}

// NewSession returns a new session.
func NewSession(id uuid.UUID, room *Room, conn *websocket.Conn) *Session {
	return &Session{
		ID:     id,
		Active: true,
		Room:   room,
		Conn:   conn,
		In:     make(chan Message),
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
		x.Conn.Close()
	}()

	for {
		messageType, message, err := x.Conn.ReadMessage()
		if err != nil {
			if err, ok := err.(*websocket.CloseError); ok {
				log.Info().Msgf("chat: close message received, code: %d, text: %s", err.Code, err.Text)
				break
			}
		}

		x.Room.Broadcast <- Message{
			Type:      messageType,
			SessionID: x.ID,
			Body:      message,
			Author:    x.ID.String(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
	}
}

func (x *Session) writePump() {
	defer x.Conn.Close()

	for message := range x.In {
		err := x.Conn.WriteMessage(int(message.Type), message.Body)
		if err != nil {
			log.Error().Err(err).Msg("chat: writing message")
			continue
		}
	}
}
