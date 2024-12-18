package service

import (
	"context"

	"github.com/google/uuid"

	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type Storer interface {
	db.Querier
	CreateRoom(ctx context.Context, arg db.CreateRoomArgs) error
	AddPlayerToRoom(ctx context.Context, arg db.AddPlayerToRoomArgs) error
	StartGame(ctx context.Context, arg db.StartGameArgs) error
	NewRound(ctx context.Context, arg db.NewRoundArgs) error
}

type Randomizer interface {
	GetNickname() string
	GetAvatar(nickname string) string
	GetRoomCode() string
	GetID() uuid.UUID
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
