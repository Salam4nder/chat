package chat

import (
	"errors"

	"github.com/Salam4nder/chat/internal/db/keyspace/chat"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
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

// MessageService defines the main message service.
// It can persist messages and communicate with NATS.
type MessageService struct {
	messageRepo chat.MessageRepository
	natsClient  *nats.Conn
}

// Message defines the message structure.
type Message struct {
	ID        uuid.UUID
	Type      int
	RoomID    string
	SessionID string
	Body      []byte
	Author    string
	Timestamp string
}

// NewMessageService returns a new instance of MessageService.
// It can persist messages and communicate with NATS.
func NewMessageService(repo chat.MessageRepository, client *nats.Conn) *MessageService {
	return &MessageService{
		messageRepo: repo,
		natsClient:  client,
	}
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
	if x.RoomID == "" {
		messageRoomIDErr = ErrMessageRoomIDInvalid
	}
	if x.SessionID == "" {
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
