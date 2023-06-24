package chat

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/olahol/melody"
)

// Room refines a chat room.
type Room struct {
	ID       uuid.UUID
	Active   bool
	Sessions []*melody.Session
	Join     chan *melody.Session
	Leave    chan *melody.Session
	In       chan []byte
	Out      chan []byte
}

// User is a user connnected to a chat room.
type User struct {
	ID      uuid.UUID
	Session *melody.Session
}

func (x *User) connectToRoom(room *Room) {
	room.Join <- x.Session
}

// NewRoom returns a new room.
func NewRoom(id uuid.UUID) *Room {
	return &Room{
		ID:       id,
		Sessions: make([]*melody.Session, 0),
		Join:     make(chan *melody.Session),
		Leave:    make(chan *melody.Session),
		In:       make(chan []byte),
		Out:      make(chan []byte),
	}
}

func GetRoom(id uuid.UUID) *Room {
	return NewRoom(id)
}

// Run runs the room.
func (x *Room) Run() {
	for {
		select {
		case session := <-x.Join:
			x.Sessions = append(x.Sessions, session)
			fmt.Println("User joined room:", x.ID)

		case session := <-x.Leave:
			session.CloseWithMsg([]byte("You have left the room."))
			fmt.Println("User left room:", x.ID)

		case message := <-x.In:
			for _, session := range x.Sessions {
				session.Write(message)
			}
		}
	}
}

type fakeCassandra struct {
	rooms map[uuid.UUID][]uuid.UUID
}

func (x *fakeCassandra) fetchConnectedRoomsForUser(id uuid.UUID) []uuid.UUID {
	return x.rooms[id]
}

// fetch all room IDs that the user is connected to from Cassandra.
// Used to reconnect the user to all rooms when they reconnect to the server.
func reconnect(id uuid.UUID) []uuid.UUID {
	return []uuid.UUID{}
}
