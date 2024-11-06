package service

import (
	"gitlab.com/hmajid2301/banterbus/internal/entities"

	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

func getLobbyPlayers(playerRows []sqlc.GetAllPlayersInRoomRow, roomCode string) entities.Lobby {
	var players []entities.LobbyPlayer
	for _, player := range playerRows {
		isHost := false
		if player.ID == player.HostPlayer {
			isHost = true
		}

		p := entities.LobbyPlayer{
			ID:       player.ID,
			Nickname: player.Nickname,
			Avatar:   string(player.Avatar),
			IsReady:  player.IsReady.Bool,
			IsHost:   isHost,
		}

		players = append(players, p)
	}

	room := entities.Lobby{
		Code:    roomCode,
		Players: players,
	}
	return room
}
