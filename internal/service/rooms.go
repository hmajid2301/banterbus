package service

import (
	"context"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type RoomService struct {
	store      Storer
	randomizer Randomizer
}

type Randomizer interface {
	GetNickname() string
	GetAvatar() []byte
}

type Storer interface {
	CreateRoom(
		ctx context.Context,
		player entities.NewPlayer,
		room entities.NewRoom,
	) (roomCode string, err error)
	AddPlayerToRoom(
		ctx context.Context,
		player entities.NewPlayer,
		roomCode string,
	) (players []sqlc.GetAllPlayersInRoomRow, err error)
	UpdateNickname(
		ctx context.Context,
		nickname string,
		playerID string,
	) (players []sqlc.GetAllPlayersInRoomRow, err error)
	UpdateAvatar(
		ctx context.Context,
		avatar []byte,
		playerID string,
	) (players []sqlc.GetAllPlayersInRoomRow, err error)
}

func NewRoomService(store Storer, randomizer Randomizer) *RoomService {
	return &RoomService{store: store, randomizer: randomizer}
}

func (r *RoomService) Create(
	ctx context.Context,
	gameName string,
	player entities.NewHostPlayer,
) (entities.Room, error) {
	newPlayer := r.getNewPlayer(player.Nickname, player.ID)

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
				Nickname: newPlayer.Nickname,
				Avatar:   string(newPlayer.Avatar),
			},
		},
	}
	return room, nil
}

func (r *RoomService) Join(
	ctx context.Context,
	roomCode string,
	playerID string,
	playerNickname string,
) (entities.Room, error) {
	newPlayer := r.getNewPlayer(playerNickname, playerID)
	// TODO: nickname exists in room
	playerRows, err := r.store.AddPlayerToRoom(ctx, newPlayer, roomCode)
	if err != nil {
		return entities.Room{}, err
	}

	room := getRoom(playerRows, roomCode)
	return room, nil
}

func (r *RoomService) getNewPlayer(playerNickname string, playerID string) entities.NewPlayer {
	nickname := playerNickname
	if playerNickname == "" {
		nickname = r.randomizer.GetNickname()
	}

	avatar := r.randomizer.GetAvatar()
	newPlayer := entities.NewPlayer{
		ID:       playerID,
		Nickname: nickname,
		Avatar:   avatar,
	}
	return newPlayer
}
