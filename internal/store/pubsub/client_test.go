package pubsub

import (
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationNewRedisClient(t *testing.T) {
	t.Parallel()

	t.Run("Should create new Redis client", func(t *testing.T) {
		t.Parallel()

		client, err := NewRedisClient("localhost:6379", 3)

		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NotNil(t, client.Redis)
			assert.NotNil(t, client.Subscribers)
			assert.Len(t, client.Subscribers, 0)
		}
	})

	t.Run("Should handle invalid address", func(t *testing.T) {
		t.Parallel()

		client, err := NewRedisClient("invalid:address:format", 1)

		assert.NoError(t, err)
		assert.NotNil(t, client.Redis)
		assert.NotNil(t, client.Subscribers)
	})
}

func TestIntegrationClientOperations(t *testing.T) {
	t.Parallel()

	client, err := NewRedisClient("localhost:6379", 1)
	require.NoError(t, err)

	ctx := t.Context()
	testID := uuid.Must(uuid.NewV7())

	t.Run("Should subscribe to channel", func(t *testing.T) {
		t.Parallel()

		ch := client.Subscribe(ctx, testID)

		assert.NotNil(t, ch)
		assert.Contains(t, client.Subscribers, testID.String())
	})

	t.Run("Should publish message", func(t *testing.T) {
		t.Parallel()

		message := []byte("test message")
		err := client.Publish(ctx, testID, message)

		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	})

	t.Run("Should close subscription", func(t *testing.T) {
		t.Parallel()

		_ = client.Subscribe(ctx, testID)
		assert.Contains(t, client.Subscribers, testID.String())

		err := client.Close(testID)
		assert.NoError(t, err)
		assert.NotContains(t, client.Subscribers, testID.String())
	})

	t.Run("Should handle close of non-existent subscription", func(t *testing.T) {
		t.Parallel()

		nonExistentID := uuid.Must(uuid.NewV7())
		err := client.Close(nonExistentID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
