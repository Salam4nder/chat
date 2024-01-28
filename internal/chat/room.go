package chat

import (
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Rooms is the main chat room registry.
var Rooms map[string]*Room

type empty struct{}

// Room defines a concurrent-safe chat room.
type Room struct {
	mu sync.Mutex

	ID        string
	Join      chan *Session
	Leave     chan *Session
	Active    bool
	Sessions  map[*Session]empty
	Broadcast chan Message
}

// NewRoom returns a new room with the given ID.
// Pass in nil to generate a new ID.
func NewRoom(roomID *string) *Room {
	if roomID == nil {
		str := uuid.NewString()
		roomID = &str
	}
	return &Room{
		ID:        *roomID,
		Join:      make(chan *Session),
		Leave:     make(chan *Session),
		Active:    true,
		Sessions:  make(map[*Session]empty),
		Broadcast: make(chan Message),
	}
}

// Run runs the main chat room engine.
// It will handle joins, leaves and room broadcasts.
func (x *Room) Run() {
	for {
		select {
		case session := <-x.Join:
			x.mu.Lock()
			x.Sessions[session] = empty{}
			x.mu.Unlock()

			log.Info().Msgf("chat: user joined room %s", x.ID)

		case session := <-x.Leave:
			close(session.In)
			session.Conn.Close()
			x.mu.Lock()
			delete(x.Sessions, session)
			x.mu.Unlock()

			log.Info().Msgf("chat: user left room %s", x.ID)

		case message := <-x.Broadcast:
			for session := range x.Sessions {
				session.In <- message
			}

			// log.Info().
			// 	Str("body", string(message.Body)).
			// 	Str("author", message.Author).
			// 	Str("room", x.ID).
			// 	Send()
		}
	}
}
