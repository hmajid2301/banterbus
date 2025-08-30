package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gofrs/uuid/v5"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	Redis       *redis.Client
	Subscribers map[string]*redis.PubSub
	mu          sync.RWMutex // Protects the Subscribers map
}

func NewRedisClient(address string, retries int) (Client, error) {
	r := redis.NewClient(&redis.Options{
		Addr:       address,
		Password:   "",
		DB:         0,
		MaxRetries: retries,
	})

	if err := redisotel.InstrumentTracing(r); err != nil {
		return Client{}, err
	}

	return Client{
		Redis:       r,
		Subscribers: map[string]*redis.PubSub{},
	}, nil
}

func (c *Client) Subscribe(ctx context.Context, id uuid.UUID) <-chan *redis.Message {
	idStr := id.String()

	c.mu.Lock()
	defer c.mu.Unlock()

	s := c.Redis.Subscribe(ctx, idStr)
	c.Subscribers[idStr] = s
	return s.Channel()
}

func (c *Client) Publish(ctx context.Context, id uuid.UUID, msg []byte) error {
	cmd := c.Redis.Publish(ctx, id.String(), msg)
	return cmd.Err()
}

func (c *Client) Close(id uuid.UUID) error {
	idStr := id.String()

	c.mu.Lock()
	pubsub, ok := c.Subscribers[idStr]
	if !ok {
		c.mu.Unlock()
		return fmt.Errorf("ID %s not found", id)
	}

	delete(c.Subscribers, idStr)
	c.mu.Unlock()

	// Safely close the pubsub with panic recovery
	var closeErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				closeErr = fmt.Errorf("panic during pubsub close: %v", r)
				slog.Error("PubSub close panic recovered", "error", r)
			}
		}()
		closeErr = pubsub.Close()
	}()

	return closeErr
}
