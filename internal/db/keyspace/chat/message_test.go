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

func TestScyllaKeyspace_CreateMessageByRoom(t *testing.T) {
	var ctx context.Context = context.Background()

	t.Run("Success", func(t *testing.T) {
		timeNow := time.Now().UTC()

		t.Cleanup(func() {
			err := TestScyllaConn.Session().Query("TRUNCATE chat.message_by_room").Exec()
			assert.NoError(t, err)
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

		query := `SELECT id, data, type, sender, room_id, time 
              FROM chat.message_by_room 
              WHERE room_id = ?`

		messages := make([]Message, 0)
		scanner := TestScyllaConn.Session().Query(query, params.RoomID).Iter().Scanner()
		for scanner.Next() {
			var message Message

			err := scanner.Scan(
				&message.ID,
				&message.Data,
				&message.Type,
				&message.Sender,
				&message.RoomID,
				&message.Time,
			)
			assert.NoError(t, err)

			messages = append(messages, message)
		}

		require.Equal(t, 1, len(messages))
		m := messages[0]
		assert.Equal(t, params.Data, m.Data)
		assert.Equal(t, params.Type, m.Type)
		assert.Equal(t, params.Sender, m.Sender)
		assert.Equal(t, params.RoomID, m.RoomID)
		assert.Equal(t, params.Timestamp.Format(time.DateTime), m.Time.Format(time.DateTime))
	})
}
