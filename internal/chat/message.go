package chat

import (
	"time"

	"github.com/google/uuid"
)

// Message defines the message structure.
type Message struct {
	ID        uuid.UUID
	Type      int
	ChannelID uuid.UUID
	Body      []byte
	Author    string
	Timestamp time.Time
}
