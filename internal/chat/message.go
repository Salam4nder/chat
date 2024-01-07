package chat

import (
	"github.com/google/uuid"
)

// Message defines the message structure.
type Message struct {
	ID        uuid.UUID
	Type      int
	RoomID    uuid.UUID
	SessionID uuid.UUID
	Body      []byte
	Author    string
	Timestamp string
}

// TypeString returns a friendly string representation of the message type.
func (x *Message) TypeString() string {
	switch x.Type {
	case 1:
		return "TextMessage"
	case 2:
		return "BinaryMessage"
	case 8:
		return "CloseMessage"
	case 9:
		return "PingMessage"
	case 10:
		return "PongMessage"
	default:
		return "Unknown"
	}
}
