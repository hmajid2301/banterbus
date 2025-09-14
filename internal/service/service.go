package service

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"gitlab.com/banterbus/banterbus/internal/store/db"
)

type Storer interface {
	db.Querier
	CreateRoom(ctx context.Context, arg db.CreateRoomArgs) error
	AddPlayerToRoom(ctx context.Context, arg db.AddPlayerToRoomArgs) error
	StartGame(ctx context.Context, arg db.StartGameArgs) error
	NewRound(ctx context.Context, arg db.NewRoundArgs) error
	NewScores(ctx context.Context, arg db.NewScoresArgs) error
	CreateQuestion(ctx context.Context, arg db.CreateQuestionArgs) (uuid.UUID, error)
}

type Randomizer interface {
	GetNickname() string
	GetAvatar(nickname string) string
	GetRoomCode() string
	GetID() (uuid.UUID, error)
	GetFibberIndex(playersLen int) int
}

func getLobbyPlayers(playerRows []db.GetAllPlayersInRoomRow, roomCode string) Lobby {
	var players []LobbyPlayer
	for _, player := range playerRows {
		isHost := false
		if player.ID == player.HostPlayer {
			isHost = true
		}

		p := LobbyPlayer{
			ID:       player.ID,
			Nickname: player.Nickname,
			Avatar:   player.Avatar,
			IsReady:  player.IsReady.Bool,
			IsHost:   isHost,
		}

		players = append(players, p)
	}

	room := Lobby{
		Code:    roomCode,
		Players: players,
	}
	return room
}

const FibberRole = "fibber"
const NormalRole = "normal"
