package chat

import (
	"context"

	"github.com/gocql/gocql"
	"github.com/rs/zerolog/log"
)

// User defines the user database model.
type User struct {
	ID     gocql.UUID
	RoomID gocql.UUID
}

// UserRepository defines a repository used to interact with users in chat rooms.
type UserRepository interface {
	// CreateUserInRoom creates an entry in the chat.user_in_room table.
	// It is used to keep track of which users are in which rooms for reconnection.
	CreateUserInRoom(ctx context.Context, params UserInRoomParams) error
	// DeleteUserInRoom deletes an entry in the chat.user_in_room table.
	// Used when a user leaves a room.
	DeleteUserInRoom(ctx context.Context, params UserInRoomParams) error
}

// ScyllaUserRepository implements the UserRepository interface.
type ScyllaUserRepository struct {
	session *gocql.Session
}

// UserInRoomParams defines the parameters used to create or delete a user in a room.
type UserInRoomParams struct {
	UserID gocql.UUID
	RoomID gocql.UUID
}

// CreateUserInRoom creates an entry in the chat.user_in_room table.
// It is used to keep track of which users are in which rooms for reconnection.
func (x *ScyllaUserRepository) CreateUserInRoom(
	ctx context.Context,
	params UserInRoomParams,
) error {
	query := `INSERT INTO chat.user_in_room 
              (user_id, room_id) 
              VALUES (?, ?)`

	if err := x.session.Query(
		query,
		params.UserID,
		params.RoomID,
	).WithContext(ctx).
		Exec(); err != nil {
		log.Error().Err(err).Msg("message: creating user in room")
		return err
	}

	return nil
}

// DeleteUserInRoom deletes an entry in the chat.user_in_room table.
func (x *ScyllaUserRepository) DeleteUserInRoom(
	ctx context.Context,
	params UserInRoomParams,
) error {
	query := `DELETE FROM chat.user_in_room 
              WHERE user_id = ? AND room_id = ?`

	if err := x.session.Query(
		query,
		params.UserID,
		params.RoomID,
	).WithContext(ctx).
		Exec(); err != nil {
		log.Error().Err(err).Msg("message: deleting user in room")
		return err
	}

	return nil
}
