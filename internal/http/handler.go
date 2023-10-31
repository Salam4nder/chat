package http

import (
	"log"
	"net/http"

	"github.com/Salam4nder/chat/internal/chat"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func handleWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// handle this better.
		log.Println(err)

		return
	}

	queryValues := r.URL.Query()

	roomID, err := uuid.Parse(queryValues.Get("roomID"))
	if err != nil {
		log.Println(err)

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
	sess := chat.NewSession(uuid.New(), chat.Rooms[roomID.String()], conn)
	go sess.Read()
	go sess.Write()
	room.Join <- sess
}