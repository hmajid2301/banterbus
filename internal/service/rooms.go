package service

import (
	"context"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/store"
)

type RoomService struct {
	store      store.Store
	Randomizer UserRandomizer
}

type UserRandomizer interface {
	GetNickname() string
	GetAvatar() []byte
}

func NewRoomService(store store.Store, randomizer UserRandomizer) *RoomService {
	return &RoomService{store: store, Randomizer: randomizer}
}

func (r *RoomService) Create(ctx context.Context, gameName string, player entities.CreateRoomPlayer) (entities.Room, error) {
	nickname := player.Nickname
	if player.Nickname == "" {
		nickname = r.Randomizer.GetNickname()
	}

	avatar := r.Randomizer.GetAvatar()
	newPlayer := entities.NewPlayer{
		ID:       player.ID,
		Nickname: nickname,
		Avatar:   avatar,
	}

	newRoom := entities.NewRoom{
		GameName: gameName,
	}
	roomCode, err := r.store.CreateRoom(ctx, newPlayer, newRoom)
	if err != nil {
		return entities.Room{}, err
	}

	room := entities.Room{
		Code: roomCode,
		Players: []entities.Player{
			{
				ID:       player.ID,
				Nickname: nickname,
				Avatar:   string(avatar),
			},
		},
	}
	return room, nil
}

func (r *RoomService) Join(ctx context.Context, roomCode string, playerID string, playerNickname string) (entities.Room, error) {
	nickname := playerNickname
	if playerNickname == "" {
		nickname = r.Randomizer.GetNickname()
	}

	avatar := r.Randomizer.GetAvatar()
	newPlayer := entities.NewPlayer{
		ID:       playerID,
		Nickname: nickname,
		Avatar:   avatar,
	}

	playerRows, err := r.store.AddPlayerToRoom(ctx, newPlayer, roomCode)
	if err != nil {
		return entities.Room{}, err
	}

	room := getRoom(playerRows, roomCode)
	return room, nil
}
