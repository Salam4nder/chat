package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

var _ MessageRepository = (*ScyllaMessageRepository)(nil)

// Message defines the message database model.
type Message struct {
	ID     gocql.UUID
	Data   []byte
	Type   string
	Sender string
	RoomID string
	Time   time.Time
}

// MessageRepository defines database methods to interact with messages.
type MessageRepository interface {
	// Session returns the underlying database session.
	Session() *gocql.Session
	// CreateMessageByRoom creates a new entry in the MessagesInRoom table.
	CreateMessageByRoom(ctx context.Context, params CreateMessageByRoomParams) error
	// ReadMessagesByRoom reads all messages from a room based on a roomID.
	ReadMessagesByRoomID(ctx context.Context, roomID string) ([]Message, error)
}

// ScyllaMessageRepository implements the MessagesRepository interface.
type ScyllaMessageRepository struct {
	session *gocql.Session
}

// NewScyllaMessageRepository creates a new instace of a ScyllaMessageRepository.
func NewScyllaMessageRepository(session *gocql.Session) *ScyllaMessageRepository {
	return &ScyllaMessageRepository{session: session}
}

// Session returns the underlying database session.
func (x *ScyllaMessageRepository) Session() *gocql.Session {
	return x.session
}

// CreateMessageByRoomParams defines the parameters to create
// a new message in a room.
type CreateMessageByRoomParams struct {
	Data      []byte
	Type      string
	Sender    string
	RoomID    string
	Timestamp time.Time
}

// CreateMessageByRoom creates a new entry in the MessageByRoom table.
func (x *ScyllaMessageRepository) CreateMessageByRoom(
	ctx context.Context,
	params CreateMessageByRoomParams,
) error {
	query := `INSERT INTO chat.message_by_room 
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
		return fmt.Errorf("message repo: creating message, %w", err)
	}

	return nil
}

// ReadMessagesByRoomID reads all messages from a room based on a roomID.
func (x *ScyllaMessageRepository) ReadMessagesByRoomID(
	ctx context.Context,
	roomID string,
) ([]Message, error) {
	query := `SELECT id, data, type, sender, room_id, time 
              FROM chat.message_by_room 
              WHERE room_id = ?`

	messages := make([]Message, 0)

	scanner := x.session.Query(
		query,
		roomID,
	).WithContext(ctx).
		Iter().
		Scanner()

	for scanner.Next() {
		var message Message

		if err := scanner.Scan(
			&message.ID,
			&message.Data,
			&message.Type,
			&message.Sender,
			&message.RoomID,
			&message.Time,
		); err != nil {
			return nil, fmt.Errorf("message repo: scanning message, %w", err)
		}
		messages = append(messages, message)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("message repo: scanner had errors, %w", err)
	}

	return messages, nil
}
