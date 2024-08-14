package ws

import (
	"bytes"
	"context"
	"fmt"

	"gitlab.com/hmajid2301/banterbus/internal/views"

	"github.com/go-viper/mapstructure/v2"
)

type message struct {
	ExtraFields map[string]interface{} `json:"-"`
	EventName   string                 `json:"event_name"`
}

type CreateRoomEvent struct {
	PlayerNickname string `mapstructure:"player_nickname"`
}

type UpdateNicknameEvent struct {
	PlayerNickname string `mapstructure:"player_nickname"`
	PlayerID       string `mapstructure:"player_id"`
}

func (s *server) handleCreateRoomEvent(ctx context.Context, client *client, message message) error {
	room := NewRoom()

	var code string
	for {
		code = s.roomRandomizer.GetRoomCode()
		if _, exists := s.rooms[code]; !exists {
			break
		}
	}
	var event CreateRoomEvent
	if err := mapstructure.Decode(message.ExtraFields, &event); err != nil {
		return fmt.Errorf("failed to decode create_room event: %w", err)
	}

	room.addClient(client)
	s.rooms[code] = room

	newRoom, err := s.roomServicer.Create(ctx, code, client.playerID, event.PlayerNickname)
	if err != nil {
		return err
	}

	go room.runRoom()

	var buf bytes.Buffer
	component := views.Room(newRoom.Code, newRoom.Players, newRoom.Players[0])
	err = component.Render(ctx, &buf)
	if err != nil {
		return err
	}

	// INFO: only one client in room, as the room has just been created by this client.
	// So doesn't matter if we broadcast the data (HTML) back or not.
	room.broadcast <- buf.Bytes()
	return nil
}

// TODO: check room state to see if possible
func (s *server) handleUpdateNicknameEvent(ctx context.Context, client *client, message message) error {
	var event UpdateNicknameEvent
	if err := mapstructure.Decode(message.ExtraFields, &event); err != nil {
		return fmt.Errorf("failed to decode update_player_nickname event: %w", err)
	}

	updatedRoom, err := s.playerServicer.UpdateNickname(ctx, event.PlayerNickname, event.PlayerID)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	clientsInRoom := s.rooms[updatedRoom.Code].clients
	for _, player := range updatedRoom.Players {
		client := clientsInRoom[player.ID]
		component := views.Room(updatedRoom.Code, updatedRoom.Players, player)
		err = component.Render(ctx, &buf)
		if err != nil {
			return err
		}
		client.messages <- buf.Bytes()

	}

	return nil
}
