package websocket

import (
	"net/http"

	"github.com/Salam4nder/chat/internal/chat"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

func HandleWS(w http.ResponseWriter, r *http.Request) {
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
	username := query.Get("name")
	if username == "" {
		log.Warn().
			Msg("websocket: parsing url query for name")

		username = "unknown"
	}

	roomIDStr := roomID.String()
	if room, exists := chat.Rooms[roomIDStr]; exists {
		session := chat.NewSession(
			chat.Rooms[roomIDStr],
			conn,
			username,
		)

		room.Join <- session

		go session.Read()
		go session.Write()

		return
	}

	room := chat.NewRoom()
	chat.Rooms[roomIDStr] = room
	go room.Run()

	session := chat.NewSession(
		chat.Rooms[roomIDStr],
		conn,
		username,
	)
	go session.Read()
	go session.Write()
	room.Join <- session
}
