package websockets

import (
	"net"
)

type client struct {
	messages   chan []byte
	connection net.Conn
	playerID   string
	locale     string
}

func newClient(conn net.Conn, playerID string) *client {
	messageChannelLength := 10
	return &client{
		playerID:   playerID,
		connection: conn,
		messages:   make(chan []byte, messageChannelLength),
	}
}
