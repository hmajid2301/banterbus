package pubsub

import (
	"context"
	"sync"

	"github.com/gofrs/uuid/v5"
	"github.com/mdobak/go-xerrors"
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

func (c Client) Subscribe(ctx context.Context, id uuid.UUID) <-chan *redis.Message {
	s := c.Redis.Subscribe(ctx, id.String())
	idStr := id.String()

	c.mu.Lock()
	c.Subscribers[idStr] = s
	c.mu.Unlock()

	return s.Channel()
}

func (c Client) Publish(ctx context.Context, id uuid.UUID, msg []byte) error {
	cmd := c.Redis.Publish(ctx, id.String(), msg)

	err := cmd.Err()
	if err != nil {
		return err
	}
	return nil
}

func (c Client) Close(id uuid.UUID) error {
	idStr := id.String()

	c.mu.Lock()
	pubsub, ok := c.Subscribers[idStr]
	if !ok {
		c.mu.Unlock()
		return xerrors.New("ID %s not found", id)
	}

	// Clean up the subscriber from the map to prevent memory leak
	delete(c.Subscribers, idStr)
	c.mu.Unlock()

	return pubsub.Close()
}
