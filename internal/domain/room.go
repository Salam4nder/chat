package domain

import (
	"fmt"

	"github.com/olahol/melody"
)

// Room refines a chat room.
type Room struct {
	ID    string
	Users map[*melody.Session]bool
	Join  chan *melody.Session
	Leave chan *melody.Session
	In    chan []byte
	Out   chan []byte
}

// NewRoom returns a new room.
func NewRoom(id string) *Room {
	return &Room{
		ID:    id,
		Users: make(map[*melody.Session]bool),
		Join:  make(chan *melody.Session),
		Leave: make(chan *melody.Session),
		In:    make(chan []byte),
		Out:   make(chan []byte),
	}
}

// Run runs the room.
func (r *Room) Run() {
	for {
		select {
		case session := <-r.Join:
			r.Users[session] = true
			fmt.Println("User joined room:", r.ID)

		case session := <-r.Leave:
			delete(r.Users, session)
			fmt.Println("User left room:", r.ID)

		case message := <-r.In:
			for session := range r.Users {
				select {
				case session.Out <- message:
				default:
					close(session.Out)
					delete(r.Users, session)
				}
			}
		}
	}
}

var rooms map[string]*Room
