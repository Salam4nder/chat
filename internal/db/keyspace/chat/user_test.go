//go:build testdb

package chat

import (
	"context"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/require"
)

func Test_CreateUserInRoom(t *testing.T) {
	timeNow := time.Now()

	params := UserInRoom{
		UserID: gocql.UUIDFromTime(timeNow),
		RoomID: gocql.UUIDFromTime(timeNow),
	}

	err := testUserRepo.CreateUserInRoom(context.Background(), params)
	require.NoError(t, err)
}
