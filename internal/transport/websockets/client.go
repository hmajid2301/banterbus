package websockets

import (
	"net"

	"github.com/gofrs/uuid/v5"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	messagesCh   <-chan *redis.Message
	connection   net.Conn
	playerID     uuid.UUID
	connectionID string
}

func newClient(conn net.Conn, playerID uuid.UUID, ch <-chan *redis.Message, connectionID string) *Client {
	return &Client{
		playerID:     playerID,
		connection:   conn,
		messagesCh:   ch,
		connectionID: connectionID,
	}
}
