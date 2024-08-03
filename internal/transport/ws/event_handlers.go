package ws

import (
	"context"
	"log"
)

type message struct {
	Data      interface{} `json:"data"`
	EventName string      `json:"event_name"`
}

func (s *server) handleRoomCreatedEvent(ctx context.Context, client *client, message message) error {
	log.Println("handle `room_created` event")
	room := NewRoom()

	var code string
	for {
		code := s.roomRandomizer.GetRoomCode()
		if _, exists := s.rooms[code]; !exists {
			break
		}
	}

	room.addClient(client)
	s.rooms[code] = room

	err := s.roomServicer.CreateRoom(ctx, code)
	if err != nil {
		return err
	}

	go room.runRoom()
	return nil
}
