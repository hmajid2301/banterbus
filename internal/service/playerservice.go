package service

import (
	"context"
	"fmt"

	"gitlab.com/hmajid2301/banterbus/internal/entities"
)

type PlayerService struct {
	store      Storer
	randomizer Randomizer
}

func NewPlayerService(store Storer, randomizer Randomizer) *PlayerService {
	return &PlayerService{store: store, randomizer: randomizer}
}

func (p *PlayerService) UpdateNickname(ctx context.Context, nickname string, playerID string) (entities.Lobby, error) {
	playerRows, err := p.store.UpdateNickname(ctx, nickname, playerID)
	if err != nil {
		return entities.Lobby{}, err
	}

	if len(playerRows) == 0 {
		return entities.Lobby{}, fmt.Errorf("no players in room")
	}

	room := getLobbyPlayers(playerRows, playerRows[0].RoomCode)
	return room, err
}

func (p *PlayerService) GenerateNewAvatar(ctx context.Context, playerID string) (entities.Lobby, error) {
	avatar := p.randomizer.GetAvatar()

	playerRows, err := p.store.UpdateAvatar(ctx, avatar, playerID)
	if err != nil {
		return entities.Lobby{}, err
	}

	if len(playerRows) == 0 {
		return entities.Lobby{}, fmt.Errorf("no players in room")
	}

	room := getLobbyPlayers(playerRows, playerRows[0].RoomCode)
	return room, err
}

func (p *PlayerService) TogglePlayerIsReady(ctx context.Context, playerID string) (entities.Lobby, error) {
	playerRows, err := p.store.ToggleIsReady(ctx, playerID)
	if err != nil {
		return entities.Lobby{}, err
	}

	if len(playerRows) == 0 {
		return entities.Lobby{}, fmt.Errorf("no players in room")
	}

	room := getLobbyPlayers(playerRows, playerRows[0].RoomCode)
	return room, err
}
