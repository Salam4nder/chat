package chat

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Session is a single chatting session in a room.
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
		x.Room.Leave <- x
		x.Conn.Close()
	}()

	for {
		messageType, message, err := x.Conn.ReadMessage()
		if err != nil {
			// handle this better
			if err.Error() == "websocket: close 1000 (normal)" {
				log.Info().Msg("chat: close message received")
				break
			}
			log.Error().Err(err).Msg("chat: reading message")
			continue
		}

		x.Room.Broadcast <- Message{
			Type:      MessageType(messageType),
			SessionID: x.ID,
			Body:      message,
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
