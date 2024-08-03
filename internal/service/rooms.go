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

func (r *RoomService) CreateRoom(ctx context.Context, roomCode string) error {
	nickname := r.Randomizer.GetNickname()
	avatar := r.Randomizer.GetAvatar()
	newPlayer := entities.NewPlayer{
		Nickname: nickname,
		Avatar:   avatar,
	}

	newRoom := entities.NewRoom{
		// TODO: don't hardcode game name
		GameName: "fibbing_it",
		RoomCode: roomCode,
	}
	return r.store.CreateRoom(ctx, newPlayer, newRoom)
}
