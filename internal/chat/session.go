package chat

import (
	"github.com/gorilla/websocket"
)

type UserSess struct {
	UserID      string
	RoomID      string
	DisplayName string
	Conn        *websocket.Conn
}
