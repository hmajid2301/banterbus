package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mdobak/go-xerrors"

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
		return Lobby{}, xerrors.New("room is not in CREATED state")
	}

	playersInRoom, err := p.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	for _, p := range playersInRoom {
		if p.Nickname == nickname {
			return Lobby{}, xerrors.New("nickname already exists")
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
		return Lobby{}, xerrors.New("room is not in CREATED state")
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
		return Lobby{}, xerrors.New("room is not in CREATED state")
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
	return err
}
