package http

import (
	"net/http"

	"github.com/Salam4nder/chat/internal/chat"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

func handleWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("http: failed to upgrade connection")
		return
	}

	queryValues := r.URL.Query()
	roomID, err := uuid.Parse(queryValues.Get("roomID"))
	if err != nil {
		log.Error().Err(err).Msg("http: failed to parse connection")
		return
	}

	if room, exists := chat.Rooms[roomID.String()]; exists {
		session := chat.NewSession(uuid.New(), chat.Rooms[roomID.String()], conn)

		room.Join <- session

		go session.Read()
		go session.Write()

		return
	}

	room := chat.NewRoom(roomID)
	chat.Rooms[roomID.String()] = room
	go room.Run()

	session := chat.NewSession(uuid.New(), chat.Rooms[roomID.String()], conn)
	go session.Read()
	go session.Write()
	room.Join <- session
}
