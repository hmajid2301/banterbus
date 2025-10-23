package service

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

var ErrPlayerNotFound = errors.New("player not found")

type PlayerService struct {
	store      Storer
	randomizer Randomizer
}

func NewPlayerService(store Storer, randomizer Randomizer) *PlayerService {
	return &PlayerService{store: store, randomizer: randomizer}
}

func (p *PlayerService) UpdateNickname(ctx context.Context, nickname string, playerID uuid.UUID) (Lobby, error) {
	result, err := p.store.UpdateNicknameWithPlayers(ctx, db.UpdateNicknameArgs{
		PlayerID: playerID,
		Nickname: nickname,
	})
	if err != nil {
		return Lobby{}, err
	}

	if len(result.Players) == 0 {
		return Lobby{}, errors.New("no players found in room")
	}

	lobby := getLobbyPlayers(result.Players, result.Players[0].RoomCode)
	return lobby, nil
}

func (p *PlayerService) GenerateNewAvatar(ctx context.Context, playerID uuid.UUID) (Lobby, error) {
	emptyName := ""
	avatar := p.randomizer.GetAvatar(emptyName)

	result, err := p.store.GenerateNewAvatarWithPlayers(ctx, db.GenerateNewAvatarArgs{
		PlayerID: playerID,
		Avatar:   avatar,
	})
	if err != nil {
		return Lobby{}, err
	}

	if len(result.Players) == 0 {
		return Lobby{}, errors.New("no players found in room")
	}

	lobby := getLobbyPlayers(result.Players, result.Players[0].RoomCode)
	return lobby, nil
}

func (p *PlayerService) TogglePlayerIsReady(ctx context.Context, playerID uuid.UUID) (Lobby, error) {
	result, err := p.store.TogglePlayerReadyWithPlayers(ctx, db.TogglePlayerIsReadyArgs{
		PlayerID: playerID,
	})
	if err != nil {
		return Lobby{}, err
	}

	if len(result.Players) == 0 {
		return Lobby{}, errors.New("no players found in room")
	}

	lobby := getLobbyPlayers(result.Players, result.Players[0].RoomCode)
	return lobby, nil
}

func (p *PlayerService) UpdateLocale(ctx context.Context, playerID uuid.UUID, newLocale string) error {
	_, err := p.store.UpdateLocale(ctx, db.UpdateLocaleParams{
		ID:     playerID,
		Locale: pgtype.Text{String: newLocale},
	})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return ErrPlayerNotFound
		}
		return err
	}
	return nil
}

func (p *PlayerService) GetPlayerByID(ctx context.Context, playerID uuid.UUID) (db.Player, error) {
	player, err := p.store.GetPlayerByID(ctx, playerID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return db.Player{}, ErrPlayerNotFound
		}
		return db.Player{}, err
	}
	return player, nil
}
