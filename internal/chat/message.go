package chat

import (
	"time"

	"github.com/google/uuid"
)

type MessageType int

const (
	// MessageTypeText is a text message.
	MessageTypeText MessageType = 1
	// MessageTypeImage is an image message.
	MessageTypeImage MessageType = 2
	// MessageTypeVideo is a video message.
	MessageTypeVideo MessageType = 3
	// MessageTypeAudio is an audio message.
	MessageTypeAudio MessageType = 4
	// MessageTypeFile is an arbitrary file message.
	MessageTypeFile MessageType = 5
)

// Message defines the message structure.
type Message struct {
	ID        uuid.UUID
	Type      MessageType
	SessionID uuid.UUID
	Body      []byte
	Author    string
	Timestamp time.Time
}
