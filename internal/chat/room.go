package chat

import (
	"log"
	"sync"

	"github.com/google/uuid"
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

			log.Println("User joined room:", x.ID)

		case session := <-x.Leave:
			session.Conn.Close()
			delete(x.Sessions, session)

			log.Println("User left room:", x.ID)

		case message := <-x.Broadcast:
			for session := range x.Sessions {
				session.In <- message
			}

		default:
			// do nothing?
		}
	}
}
