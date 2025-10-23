package service

import (
	"github.com/gofrs/uuid/v5"

	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

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
