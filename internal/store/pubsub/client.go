package pubsub

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdobak/go-xerrors"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	Redis       *redis.Client
	Subscribers map[string]*redis.PubSub
}

func NewRedisClient(address string) (Client, error) {
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
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
	c.Subscribers[id.String()] = s
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
	pubsub, ok := c.Subscribers[id.String()]
	if !ok {
		return xerrors.New("ID %s not found", id)
	}

	return pubsub.Close()
}
