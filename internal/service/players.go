package service

import (
	"context"
	"fmt"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
	"gitlab.com/hmajid2301/banterbus/internal/store"
)

type PlayerService struct {
	store store.Store
}

func NewPlayerService(store store.Store) *PlayerService {
	return &PlayerService{store: store}
}

func (p *PlayerService) UpdateNickname(ctx context.Context, nickname string, playerID string) (entities.Room, error) {
	playerRows, err := p.store.UpdateNickname(ctx, nickname, playerID)
	if err != nil {
		return entities.Room{}, err
	}

	var players []entities.Player
	for _, player := range playerRows {
		p := entities.Player{
			ID:       player.ID,
			Nickname: player.Nickname,
			Avatar:   string(player.Avatar),
		}

		players = append(players, p)
	}

	if len(players) == 0 {
		return entities.Room{}, fmt.Errorf("no players in room, playerID: %s", playerID)
	}

	room := entities.Room{
		Code:    playerRows[0].RoomCode,
		Players: players,
	}

	return room, err
}
