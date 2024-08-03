package ws

import "net"

type client struct {
	connection net.Conn
	messages   chan []byte
}

func NewClient(conn net.Conn) *client {
	return &client{
		connection: conn,
		messages:   make(chan []byte, 10),
	}
}
