//go:build testdb

package chat

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// func TestScyllaRepository_Create(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		params      *Message
// 		want        *Message
// 		wantErr     bool
// 		requiredErr error
// 	}{
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := TestScyllaConn.Create()

// 		})
// 	}
// }
var ctx context.Context = context.Background()

func TestScyllaKeyspace_CreateMessageByRoom(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		timeNow := time.Now().UTC()

		t.Cleanup(func() {
			// err := TestScyllaConn.Session().ExecStmt("DELETE from message.message_by_room")
			// assert.NoError(t, err)
		})

		params := CreateMessageByRoomParam{
			Data:      []byte("test"),
			Type:      "text",
			Sender:    "test_sender",
			RoomID:    uuid.New().String(),
			Timestamp: timeNow,
		}

		err := TestScyllaConn.CreateMessageByRoom(ctx, params)
		require.NoError(t, err)

		messages, err := TestScyllaConn.ReadMessagesByRoom(ctx, params.RoomID)
		require.NoError(t, err)
		require.Equal(t, 1, len(messages))

		assert.Equal(t, params.Data, messages[0].Data)
		assert.Equal(t, params.Type, messages[0].Type)
		assert.Equal(t, params.Sender, messages[0].Sender)
		assert.Equal(t, params.RoomID, messages[0].RoomID)
		assert.Equal(t, params.Timestamp.Format(time.DateTime), messages[0].Time.Format(time.DateTime))
	})
}
