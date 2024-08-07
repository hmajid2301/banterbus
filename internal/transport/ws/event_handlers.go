package ws

import (
	"bytes"
	"context"
	"log"

	"gitlab.com/hmajid2301/banterbus/internal/views"
)

type message struct {
	Data      interface{} `json:"data"`
	EventName string      `json:"event_name"`
}

func (s *server) handleRoomCreatedEvent(ctx context.Context, client *client, message message) ([]byte, error) {
	log.Println("handle `room_created` event")
	room := NewRoom()

	var code string
	for {
		code = s.roomRandomizer.GetRoomCode()
		if _, exists := s.rooms[code]; !exists {
			break
		}
	}

	room.addClient(client)
	s.rooms[code] = room

	newRoom, err := s.roomServicer.CreateRoom(ctx, code)
	if err != nil {
		return nil, err
	}

	go room.runRoom()

	comp := views.Room(newRoom.Code, newRoom.Players)

	var buf bytes.Buffer
	err = comp.Render(ctx, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
