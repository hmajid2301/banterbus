package ws

import "net"

type client struct {
	connection *websocket.Conn
	messages   chan []byte
}

func NewClient(conn *websocket.Conn) *client {
	return &client{
		connection: conn,
		messages:   make(chan []byte, 10),
	}
}
