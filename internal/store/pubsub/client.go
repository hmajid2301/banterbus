package pubsub

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	Redis       *redis.Client
	Subscribers map[string]*redis.PubSub
}

func NewRedisClient(address string) Client {
	r := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})
	return Client{
		Redis:       r,
		Subscribers: map[string]*redis.PubSub{},
	}
}

func (c Client) Subscribe(ctx context.Context, id string) <-chan *redis.Message {
	s := c.Redis.Subscribe(ctx, id)
	c.Subscribers[id] = s
	return s.Channel()
}

func (c Client) Publish(ctx context.Context, id string, msg []byte) error {
	cmd := c.Redis.Publish(ctx, id, msg)

	err := cmd.Err()
	if err != nil {
		return err
	}
	return nil
}

func (c Client) Close(id string) error {
	pubsub, ok := c.Subscribers[id]
	if !ok {
		return fmt.Errorf("ID %s not found", id)
	}

	return pubsub.Close()
}
