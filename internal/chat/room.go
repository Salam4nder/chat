package chat

import (
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Rooms is the main chat room registry.
var Rooms map[string]*Room

type empty struct{}

// Room refines a chat room.
type Room struct {
	sync.Mutex

	ID        uuid.UUID
	Join      chan *Session
	Leave     chan *Session
	Active    bool
	Sessions  map[*Session]empty
	Broadcast chan Message
}

// NewRoom returns a new room.
func NewRoom(id uuid.UUID) *Room {
	if id == uuid.Nil {
		id = uuid.New()
	}

	return &Room{
		ID:        id,
		Join:      make(chan *Session),
		Leave:     make(chan *Session),
		Active:    true,
		Sessions:  make(map[*Session]empty),
		Broadcast: make(chan Message),
	}
}

// Run runs the room.
func (x *Room) Run() {
	for {
		select {
		case session := <-x.Join:
			x.Lock()
			x.Sessions[session] = empty{}
			x.Unlock()

			log.Info().Msgf("chat: user joined room %s", x.ID.String())

		case session := <-x.Leave:
			session.Conn.Close()
			delete(x.Sessions, session)

			log.Info().Msgf("chat: user left room %s", x.ID.String())

		case message := <-x.Broadcast:
			for session := range x.Sessions {
				session.In <- message
			}

			log.Info().Msgf("chat: %s broadcasted to room %s", string(message.Body), x.ID.String())
		}
	}
}
