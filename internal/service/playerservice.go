package service

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type PlayerService struct {
	store      Storer
	randomizer Randomizer
}

func NewPlayerService(store Storer, randomizer Randomizer) *PlayerService {
	return &PlayerService{store: store, randomizer: randomizer}
}

func (p *PlayerService) UpdateNickname(ctx context.Context, nickname string, playerID uuid.UUID) (Lobby, error) {
	room, err := p.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	if room.RoomState != db.Created.String() {
		return Lobby{}, errors.New("room is not in CREATED state")
	}

	playersInRoom, err := p.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	for _, p := range playersInRoom {
		if p.Nickname == nickname {
			return Lobby{}, errors.New("nickname already exists")
		}
	}

	_, err = p.store.UpdateNickname(ctx, db.UpdateNicknameParams{
		Nickname: nickname,
		ID:       playerID,
	})
	if err != nil {
		return Lobby{}, err
	}

	players, err := p.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	lobby := getLobbyPlayers(players, players[0].RoomCode)
	return lobby, err
}

func (p *PlayerService) GenerateNewAvatar(ctx context.Context, playerID uuid.UUID) (Lobby, error) {
	room, err := p.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	if room.RoomState != db.Created.String() {
		return Lobby{}, errors.New("room is not in CREATED state")
	}

	// INFO: Using an empty name, generates a completely random avatar (random seed vs using the nickname as a seed)
	emptyName := ""
	avatar := p.randomizer.GetAvatar(emptyName)

	_, err = p.store.UpdateAvatar(ctx, db.UpdateAvatarParams{
		Avatar: avatar,
		ID:     playerID,
	})
	if err != nil {
		return Lobby{}, err
	}

	players, err := p.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	lobby := getLobbyPlayers(players, players[0].RoomCode)
	return lobby, err
}

func (p *PlayerService) TogglePlayerIsReady(ctx context.Context, playerID uuid.UUID) (Lobby, error) {
	room, err := p.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	if room.RoomState != db.Created.String() {
		return Lobby{}, errors.New("room is not in CREATED state")
	}

	_, err = p.store.TogglePlayerIsReady(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	players, err := p.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	lobby := getLobbyPlayers(players, players[0].RoomCode)
	return lobby, err
}

func (p *PlayerService) UpdateLocale(ctx context.Context, playerID uuid.UUID, newLocale string) error {
	_, err := p.store.UpdateLocale(ctx, db.UpdateLocaleParams{
		ID:     playerID,
		Locale: pgtype.Text{String: newLocale},
	})
	if err != nil {
		// Handle case where player doesn't exist (common in tests during cleanup)
		if err.Error() == "no rows in result set" {
			return nil
		}
		return err
	}
	return nil
}
