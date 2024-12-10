package service

import (
	"context"

	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type Storer interface {
	db.Querier
	CreateRoom(ctx context.Context, arg db.CreateRoomParams) error
	AddPlayerToRoom(ctx context.Context, arg db.AddPlayerToRoomArgs) error
	StartGame(ctx context.Context, arg db.StartGameArgs) error
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
			Avatar:   string(player.Avatar),
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
