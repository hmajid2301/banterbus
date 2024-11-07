package websockets

import (
	"net"

	"github.com/flexstack/uuid"
)

type client struct {
	messages   chan []byte
	connection net.Conn
	playerID   string
	locale     string
}

func newClient(conn net.Conn) *client {
	u := uuid.Must(uuid.NewV7())

	messageChannelLength := 10
	return &client{
		playerID:   u.String(),
		connection: conn,
		messages:   make(chan []byte, messageChannelLength),
	}
}
