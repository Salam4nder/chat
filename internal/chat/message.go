package chat

import (
	"github.com/google/uuid"
)

// Message defines the message structure.
type Message struct {
	ID        uuid.UUID
	ChannelID uuid.UUID
	Body      interface{}
	Author    string
}
