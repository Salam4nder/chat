package chat

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrMessageIDInvalid        = errors.New("message ID invalid")
	ErrMessageTypeInvalid      = errors.New("message type invalid")
	ErrMessageRoomIDInvalid    = errors.New("message room ID invalid")
	ErrMessageSessionIDInvalid = errors.New("message session ID invalid")
	ErrMessageBodyInvalid      = errors.New("message body invalid")
	ErrMessageAuthorInvalid    = errors.New("message author invalid")
	ErrMessageTimestampInvalid = errors.New("message timestamp invalid")
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

// Valid returns nil if all the fields of Message are valid.
func (x *Message) Valid() error {
	var (
		messageIDErr        error
		messageTypeErr      error
		messageRoomIDErr    error
		messageSessionIDErr error
		messageBodyErr      error
		messageAuthorErr    error
		messageTimestampErr error
	)

	if x.ID == uuid.Nil {
		messageIDErr = ErrMessageIDInvalid
	}
	if x.Type == 0 || x.Type > 10 {
		messageTypeErr = ErrMessageTypeInvalid
	}
	if x.RoomID == uuid.Nil {
		messageRoomIDErr = ErrMessageRoomIDInvalid
	}
	if x.SessionID == uuid.Nil {
		messageSessionIDErr = ErrMessageSessionIDInvalid
	}
	if len(x.Body) == 0 {
		messageBodyErr = ErrMessageBodyInvalid
	}
	if x.Author == "" {
		messageAuthorErr = ErrMessageAuthorInvalid
	}
	if x.Timestamp == "" {
		messageTimestampErr = ErrMessageTimestampInvalid
	}

	return errors.Join(
		messageIDErr,
		messageTypeErr,
		messageRoomIDErr,
		messageSessionIDErr,
		messageBodyErr,
		messageAuthorErr,
		messageTimestampErr,
	)
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
