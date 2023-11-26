package message

import (
	"errors"
	"time"

	"github.com/gocql/gocql"
	"github.com/rs/zerolog/log"
)

var (
	// ErrMessageNotFound is returned when a message is not found in the database.
	ErrMessageNotFound            = errors.New("repository: message not found")
	_                  Repository = (*RepositoryImpl)(nil)
)

// Response defines a message model response from the database.
type Response struct {
	ID        gocql.UUID `json:"id"`
	Type      string     `json:"type"`
	Data      string     `json:"data"`
	Sender    string     `json:"sender"`
	Recipient string     `json:"recipient"`
	Timestamp time.Time  `json:"time"`
}

// Repository defines the db methods to interact with the message table.
type Repository interface {
	// Create a new message in the database and return it.
	Create(params CreateParams) (*Response, error)
	// Read a message from the database and return it.
	Read(id gocql.UUID) (*Response, error)
}

// RepositoryImpl implements the Repository interface.
type RepositoryImpl struct {
	session *gocql.Session
}

// NewRepository creates a new message repository.
func NewRepository(session *gocql.Session) *RepositoryImpl {
	return &RepositoryImpl{session: session}
}

// CreateParams defines the parameters to create a new message.
type CreateParams struct {
	Type      string
	Data      string
	Sender    string
	Recipient string
	Timestamp time.Time
}

// Create a new message in the database and return it.
func (x *RepositoryImpl) Create(params CreateParams) (*Response, error) {
	query := `INSERT INTO messages (id, type, data, sender, recipient, timestamp) 
              VALUES (?, ?, ?, ?, ?, ?)`

	var message Response
	if err := x.session.Query(
		query,
		gocql.TimeUUID(),
		params.Type,
		params.Data,
		params.Sender,
		params.Recipient,
		params.Timestamp,
	).Scan(&message); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, ErrMessageNotFound
		}
		log.Error().Err(err).Msg("repository: failed to create message")
	}
	return &message, nil
}

// Read a message from the database and return it.
func (x *RepositoryImpl) Read(id gocql.UUID) (*Response, error) {
	query := `SELECT id, type, data, sender, recipient, timestamp FROM messages WHERE id = ?`

	var message Response
	if err := x.session.Query(query, id).Scan(
		&message.ID,
		&message.Type,
		&message.Data,
		&message.Sender,
		&message.Recipient,
		&message.Timestamp,
	); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, ErrMessageNotFound
		}
		log.Error().Err(err).Msg("repository: failed to read message")
	}
	return &message, nil
}
