package service

import (
	"context"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type LobbyService struct {
	store      Storer
	randomizer Randomizer
}

type Randomizer interface {
	GetNickname() string
	GetAvatar() []byte
}

type Storer interface {
	CreateRoom(ctx context.Context, player entities.NewPlayer, room entities.NewRoom) (roomCode string, err error)
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
	UpdateAvatar(ctx context.Context, avatar []byte, playerID string) (players []sqlc.GetAllPlayersInRoomRow, err error)
	ToggleIsReady(ctx context.Context, playerID string) (players []sqlc.GetAllPlayersInRoomRow, err error)
	StartGame(ctx context.Context, roomCode string, playerID string) (players []sqlc.GetAllPlayersInRoomRow, err error)
}

func NewLobbyService(store Storer, randomizer Randomizer) *LobbyService {
	return &LobbyService{store: store, randomizer: randomizer}
}

func (r *LobbyService) Create(
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

func (r *LobbyService) Join(
	ctx context.Context,
	roomCode string,
	playerID string,
	playerNickname string,
) (entities.Room, error) {
	newPlayer := r.getNewPlayer(playerNickname, playerID)
	playerRows, err := r.store.AddPlayerToRoom(ctx, newPlayer, roomCode)
	if err != nil {
		return entities.Room{}, err
	}

	room := getRoom(playerRows, roomCode)
	return room, nil
}

func (r *LobbyService) Start(ctx context.Context, roomCode string, playerID string) (entities.Room, error) {
	playerRows, err := r.store.StartGame(ctx, roomCode, playerID)
	if err != nil {
		return entities.Room{}, err
	}

	room := getRoom(playerRows, roomCode)
	return room, nil
}

func (r *LobbyService) getNewPlayer(playerNickname string, playerID string) entities.NewPlayer {
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
