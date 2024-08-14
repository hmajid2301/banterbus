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

// TODO: maybe refactor player info into a struct.
func (r *RoomService) Create(ctx context.Context, roomCode string, playerID string, playerNickname string) (entities.Room, error) {
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

	newRoom := entities.NewRoom{
		// TODO: don't hardcode game name
		GameName: "fibbing_it",
		RoomCode: roomCode,
	}
	err := r.store.CreateRoom(ctx, newPlayer, newRoom)
	if err != nil {
		return entities.Room{}, err
	}

	room := entities.Room{
		Code: roomCode,
		Players: []entities.Player{
			{
				ID:       playerID,
				Nickname: nickname,
				Avatar:   string(avatar),
			},
		},
	}
	return room, nil
}
