package websockets

import (
	"net"

	"github.com/redis/go-redis/v9"
)

type client struct {
	messagesCh <-chan *redis.Message
	connection net.Conn
	playerID   string
	locale     string
}

func newClient(conn net.Conn, playerID string, ch <-chan *redis.Message) *client {
	return &client{
		playerID:   playerID,
		connection: conn,
		messagesCh: ch,
	}
}
