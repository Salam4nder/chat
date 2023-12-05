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
