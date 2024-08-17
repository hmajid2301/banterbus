package ws

import (
	"bytes"
	"context"
	"fmt"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/views"

	"github.com/go-viper/mapstructure/v2"
)

// TODO: refactor to anotther package
type RoomServicer interface {
	Create(ctx context.Context, gameName string, player entities.CreateRoomPlayer) (entities.Room, error)
	Join(ctx context.Context, roomCode string, playerID string, playerNickname string) (entities.Room, error)
}

type PlayerServicer interface {
	UpdateNickname(ctx context.Context, nickname string, playerID string) (entities.Room, error)
	GenerateNewAvatar(ctx context.Context, playerID string) (entities.Room, error)
}

type CreateRoomEvent struct {
	GameName       string `mapstructure:"game_name"`
	PlayerNickname string `mapstructure:"player_nickname"`
}

type JoinRoomEvent struct {
	PlayerNickname string `mapstructure:"player_nickname"`
	RoomCode       string `mapstructure:"room_code"`
}

type UpdateNicknameEvent struct {
	PlayerNickname string `mapstructure:"player_nickname"`
	PlayerID       string `mapstructure:"player_id"`
}

type GenerateNewAvatarEvent struct {
	PlayerID string `mapstructure:"player_id"`
}

func (s *subscriber) handleCreateRoomEvent(ctx context.Context, client *client, message message) error {
	var event CreateRoomEvent
	if err := mapstructure.Decode(message.ExtraFields, &event); err != nil {
		return fmt.Errorf("failed to decode create_room event: %w", err)
	}

	newPlayer := entities.CreateRoomPlayer{
		ID:       client.playerID,
		Nickname: event.PlayerNickname,
	}
	newRoom, err := s.roomServicer.Create(ctx, event.GameName, newPlayer)
	if err != nil {
		return err
	}

	room := NewRoom()

	room.addClient(client)
	s.rooms[newRoom.Code] = room

	go room.runRoom()

	err = s.updateClients(ctx, newRoom)
	return err
}

func (s *subscriber) handleJoinRoomEvent(ctx context.Context, client *client, message message) error {
	var event JoinRoomEvent
	if err := mapstructure.Decode(message.ExtraFields, &event); err != nil {
		return fmt.Errorf("failed to decode join_room event: %w", err)
	}

	room, ok := s.rooms[event.RoomCode]
	if !ok {
		return fmt.Errorf("room with code %s does not exist", event.RoomCode)
	}
	room.addClient(client)

	updatedRoom, err := s.roomServicer.Join(ctx, event.RoomCode, client.playerID, event.PlayerNickname)
	if err != nil {
		return err
	}

	err = s.updateClients(ctx, updatedRoom)
	return err
}

func (s *subscriber) handleUpdateNicknameEvent(ctx context.Context, client *client, message message) error {
	var event UpdateNicknameEvent
	if err := mapstructure.Decode(message.ExtraFields, &event); err != nil {
		return fmt.Errorf("failed to decode update_player_nickname event: %w", err)
	}

	updatedRoom, err := s.playerServicer.UpdateNickname(ctx, event.PlayerNickname, event.PlayerID)
	if err != nil {
		return err
	}

	err = s.updateClients(ctx, updatedRoom)
	return err
}

func (s *subscriber) handleGenerateNewAvatarEvent(ctx context.Context, client *client, message message) error {
	var event GenerateNewAvatarEvent
	if err := mapstructure.Decode(message.ExtraFields, &event); err != nil {
		return fmt.Errorf("failed to decode generate_new_avatar event: %w", err)
	}

	updatedRoom, err := s.playerServicer.GenerateNewAvatar(ctx, event.PlayerID)
	if err != nil {
		return err
	}

	err = s.updateClients(ctx, updatedRoom)
	return err
}

func (s *subscriber) updateClients(ctx context.Context, updatedRoom entities.Room) error {
	var buf bytes.Buffer
	clientsInRoom := s.rooms[updatedRoom.Code].clients
	for _, player := range updatedRoom.Players {
		client := clientsInRoom[player.ID]
		component := views.Room(updatedRoom.Code, updatedRoom.Players, player)
		err := component.Render(ctx, &buf)
		if err != nil {
			return err
		}
		client.messages <- buf.Bytes()

	}
	return nil
}
