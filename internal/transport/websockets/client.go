package websockets

import (
	"net"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type client struct {
	messagesCh <-chan *redis.Message
	connection net.Conn
	playerID   uuid.UUID
}

func newClient(conn net.Conn, playerID uuid.UUID, ch <-chan *redis.Message) *client {
	return &client{
		playerID:   playerID,
		connection: conn,
		messagesCh: ch,
	}
}
