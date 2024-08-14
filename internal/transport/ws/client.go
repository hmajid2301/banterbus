package ws

import (
	"net"

	"github.com/flexstack/uuid"
)

type client struct {
	messages   chan []byte
	connection net.Conn
	playerID   string
}

func NewClient(conn net.Conn) *client {
	u := uuid.Must(uuid.NewV7())

	return &client{
		playerID:   u.String(),
		connection: conn,
		messages:   make(chan []byte, 10),
	}
}
