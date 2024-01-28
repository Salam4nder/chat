package websocket

import (
	"net/http"

	"github.com/Salam4nder/chat/internal/chat"
	"github.com/Salam4nder/chat/internal/event"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	registry *event.Registry
}

// NewHandler creates a new websocket handler.
func NewHandler(registry *event.Registry) *Handler {
	return &Handler{registry: registry}
}

// HandleConnect handles a new /chat connection.
// It hanldes websocket upgrades and notifies about connection details.
func (x *Handler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket: upgrading connection")
		return
	}

	if r.URL == nil {
		log.Error().Msg("websocket: url is nil")
		return
	}
	query := r.URL.Query()
	roomID, err := uuid.Parse(query.Get("roomID"))
	if err != nil {
		log.Error().
			Err(err).
			Msg("websocket: parsing url query for roomID")
		return
	}
	userID, err := uuid.Parse(query.Get("userID"))
	if err != nil {
		log.Error().
			Err(err).
			Msg("websocket: parsing url query for userID")
		return
	}
	username := query.Get("name")
	if username == "" {
		log.Warn().
			Msg("websocket: parsing url query for username")

		username = "unknown"
	}

	if err := x.registry.Publish(
		event.New(chat.SessionConnectedEvent, chat.SessionConnectedPayload{
			UserID:   userID.String(),
			RoomID:   roomID.String(),
			Username: username,
			Conn:     conn,
		}),
	); err != nil {
		log.Error().
			Err(err).
			Msg("websocket: publishing session connected event")
	}
}
