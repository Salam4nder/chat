package message

import (
	"context"
	"errors"
	"time"

	"github.com/gocql/gocql"
	"github.com/rs/zerolog/log"
)

var (
	// ErrMessageNotFound is returned when a message is not found in the database.
	ErrMessageNotFound = errors.New("message: message not found")

	_ Keyspace = (*ScyllaKeyspace)(nil)
)

// Response defines a message response from the database.
type Response struct {
	ID     gocql.UUID
	Data   []byte
	Type   string
	Sender string
	RoomID string
	Time   time.Time
}

// Keyspace defines database methods to interact with messages.
type Keyspace interface {
	// Session returns the underlying database session.
	Session() *gocql.Session
	// CreateMessageByRoom creates a new entry in the MessagesInRoom table.
	CreateMessageByRoom(ctx context.Context, params CreateMessageByRoomParam) error
	// ReadMessagesByRoom reads all messages from a room based on a roomID.
	ReadMessagesByRoom(ctx context.Context, roomID string) ([]Response, error)
}

// ScyllaKeyspace implements the MessagesKeyspace interface.
type ScyllaKeyspace struct {
	session *gocql.Session
}

// NewKeyspace creates a new instace of a message.Keyspace.
func NewKeyspace(session *gocql.Session) *ScyllaKeyspace {
	return &ScyllaKeyspace{session: session}
}

// Session returns the underlying database session.
func (x *ScyllaKeyspace) Session() *gocql.Session {
	return x.session
}

// CreateMessageByRoomParam defines the parameters to create
// a new message in a room.
type CreateMessageByRoomParam struct {
	Data      []byte
	Type      string
	Sender    string
	RoomID    string
	Timestamp time.Time
}

// CreateMessageByRoom creates a new entry in the MessageByRoom table.
func (x *ScyllaKeyspace) CreateMessageByRoom(
	ctx context.Context,
	params CreateMessageByRoomParam,
) error {
	query := `INSERT INTO message.message_by_room 
              (id, data, type, sender, room_id, time) 
              VALUES (?, ?, ?, ?, ?, ?)`

	if err := x.session.Query(
		query,
		gocql.UUIDFromTime(time.Now()),
		params.Data,
		params.Type,
		params.Sender,
		params.RoomID,
		params.Timestamp,
	).WithContext(ctx).
		Exec(); err != nil {
		log.Error().Err(err).Msg("message: creating message")
		return err
	}

	return nil
}

// ReadMessagesByRoom reads all messages from a room based on a roomID.
func (x *ScyllaKeyspace) ReadMessagesByRoom(ctx context.Context, roomID string) ([]Response, error) {
	query := `SELECT id, data, type, sender, room_id, time 
              FROM message.message_by_room 
              WHERE room_id = ?`

	messages := make([]Response, 0)

	scanner := x.session.Query(
		query,
		roomID,
	).WithContext(ctx).
		Iter().
		Scanner()

	for scanner.Next() {
		var message Response

		if err := scanner.Scan(
			&message.ID,
			&message.Data,
			&message.Type,
			&message.Sender,
			&message.RoomID,
			&message.Time,
		); err != nil {
			log.Error().Err(err).Msg("message: scanning message")
			return nil, err
		}
		messages = append(messages, message)
	}

	if err := scanner.Err(); err != nil {
		log.Error().Err(err).Msg("message: scanner had errors")
		return nil, err
	}

	return messages, nil
}
