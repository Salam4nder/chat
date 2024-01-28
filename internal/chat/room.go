package chat

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/Salam4nder/chat/internal/event"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type empty struct{}

type Rooms map[string]*Room

// ChatRomoms is the main chat room registry.
var ChatRomoms Rooms

func (x *Rooms) Run(m chan *nats.Msg, interrupt chan os.Signal) {
	for {
		select {
		case msg := <-m:
			if msg == nil {
				return
			}
			var message Message
			if err := gob.NewDecoder(bytes.NewReader(msg.Data)).
				Decode(&message); err != nil {
				log.Error().
					Err(err).
					Msg("failed to decode message")
			}
			if room, ok := (*x)[message.RoomID]; ok {
				room.broadcast(message)
			}

		case i := <-interrupt:
			for _, room := range *x {
				room.interrupt <- i
			}
			return
		}
	}
}

// Room defines a concurrent-safe chat room.
type Room struct {
	mu sync.Mutex

	ID       string
	Join     chan *UserSess
	Leave    chan *UserSess
	Sessions map[*UserSess]empty

	interrupt     chan os.Signal
	eventRegistry *event.Registry
}

// NewRoom returns a new room with the given ID.
// Pass in nil to generate a new ID.
func NewRoom(roomID *string, registry *event.Registry) *Room {
	if roomID == nil {
		str := uuid.NewString()
		roomID = &str
	}
	if registry == nil {
		registry = event.NewRegistry()
	}
	return &Room{
		ID:            *roomID,
		Join:          make(chan *UserSess),
		Leave:         make(chan *UserSess),
		Sessions:      make(map[*UserSess]empty),
		eventRegistry: registry,
	}
}

// Run runs the main chat room engine.
// It will handle joins, leaves and writes.
func (x *Room) Run(ctx context.Context) {
	for {
		select {
		case session := <-x.Join:
			x.mu.Lock()
			x.Sessions[session] = empty{}
			x.mu.Unlock()
			go x.serveConn(ctx, session)
			log.Info().Msgf("chat: user joined room %s", x.ID)

		case session := <-x.Leave:
			session.Conn.Close()
			x.mu.Lock()
			delete(x.Sessions, session)
			x.mu.Unlock()
			log.Info().Msgf("chat: user left room %s", x.ID)

		case <-x.interrupt:
			log.Info().Msgf("chat: room %s interrupted", x.ID)
			return
		}
	}
}

func (x *Room) serveConn(ctx context.Context, sess *UserSess) {
	for {
		mType, m, err := sess.Conn.ReadMessage()
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
			RoomID:    sess.RoomID,
			SessionID: sess.UserID,
			Body:      m,
			Author:    sess.DisplayName,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		if err := x.eventRegistry.Publish(
			ctx,
			event.New(MessageCreatedInRoomEvent, message),
		); err != nil {
			log.Error().Err(err).Msg("chat: publishing message")
		}
	}
}

func (x *Room) broadcast(m Message) {
	for sess := range x.Sessions {
		err := sess.Conn.WriteMessage(m.Type, m.Body)
		if err != nil {
			log.Error().Err(err).Msg("chat: writing message")
		}
	}
}
