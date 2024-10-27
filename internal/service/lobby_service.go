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
	StartGame(ctx context.Context, roomCode string, playerID string) (gameState entities.GameState, err error)
	KickPlayer(
		ctx context.Context,
		roomCode string,
		playerID string,
		playerNicknameToKick string,
	) (players []sqlc.GetAllPlayersInRoomRow, playerToKickID string, err error)
}

func NewLobbyService(store Storer, randomizer Randomizer) *LobbyService {
	return &LobbyService{store: store, randomizer: randomizer}
}

func (r *LobbyService) Create(
	ctx context.Context,
	gameName string,
	player entities.NewHostPlayer,
) (entities.Lobby, error) {
	newPlayer := r.getNewPlayer(player.Nickname, player.ID)

	newRoom := entities.NewRoom{
		GameName: gameName,
	}
	roomCode, err := r.store.CreateRoom(ctx, newPlayer, newRoom)
	if err != nil {
		return entities.Lobby{}, err
	}

	lobby := entities.Lobby{
		Code: roomCode,
		Players: []entities.LobbyPlayer{
			{
				ID:       player.ID,
				Nickname: newPlayer.Nickname,
				Avatar:   string(newPlayer.Avatar),
			},
		},
	}
	return lobby, nil
}

func (r *LobbyService) Join(
	ctx context.Context,
	roomCode string,
	playerID string,
	playerNickname string,
) (entities.Lobby, error) {
	newPlayer := r.getNewPlayer(playerNickname, playerID)
	playerRows, err := r.store.AddPlayerToRoom(ctx, newPlayer, roomCode)
	if err != nil {
		return entities.Lobby{}, err
	}

	lobby := getLobbyPlayers(playerRows, roomCode)
	return lobby, nil
}

func (r *LobbyService) Start(
	ctx context.Context,
	roomCode string,
	playerID string,
) (entities.GameState, error) {
	gameState, err := r.store.StartGame(ctx, roomCode, playerID)
	if err != nil {
		return entities.GameState{}, err
	}

	return gameState, nil
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

func (r *LobbyService) KickPlayer(
	ctx context.Context,
	roomCode string,
	playerID string,
	playerNicknameToKick string,
) (entities.Lobby, string, error) {
	playerRows, playerNicknameToKick, err := r.store.KickPlayer(ctx, roomCode, playerID, playerNicknameToKick)
	if err != nil {
		return entities.Lobby{}, "", err
	}

	lobby := getLobbyPlayers(playerRows, roomCode)
	return lobby, playerNicknameToKick, nil
}
