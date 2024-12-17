package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

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

	if room.RoomState != db.ROOMSTATE_CREATED.String() {
		return Lobby{}, fmt.Errorf("room is not in CREATED state")
	}

	playersInRoom, err := p.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	for _, p := range playersInRoom {
		if p.Nickname == nickname {
			return Lobby{}, fmt.Errorf("nickname already exists")
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

	if room.RoomState != db.ROOMSTATE_CREATED.String() {
		return Lobby{}, fmt.Errorf("room is not in CREATED state")
	}

	avatar := p.randomizer.GetAvatar()

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

	if room.RoomState != db.ROOMSTATE_CREATED.String() {
		return Lobby{}, fmt.Errorf("room is not in CREATED state")
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

// TODO: move these to their own service file don't really belong
func (p *PlayerService) GetRoomState(ctx context.Context, playerID uuid.UUID) (db.RoomState, error) {
	room, err := p.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return db.ROOMSTATE_CREATED, err
	}

	roomState, err := db.RoomStateFromString(room.RoomState)
	return roomState, err
}

func (p *PlayerService) GetLobby(ctx context.Context, playerID uuid.UUID) (Lobby, error) {
	players, err := p.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	room := getLobbyPlayers(players, players[0].RoomCode)
	return room, err
}
func (p *PlayerService) GetGameState(ctx context.Context, playerID uuid.UUID) (db.GameStateEnum, error) {
	game, err := p.store.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return db.GAMESTATE_FIBBING_IT_QUESTION, err
	}

	gameState, err := db.GameStateFromString(game.State)
	return gameState, err
}

func (p *PlayerService) GetQuestionState(ctx context.Context, playerID uuid.UUID) (QuestionState, error) {
	g, err := p.store.GetCurrentQuestionByPlayerID(ctx, playerID)
	if err != nil {
		return QuestionState{}, err
	}

	players := []PlayerWithRole{
		{
			ID:            g.PlayerID,
			Role:          g.Role.String,
			Question:      g.Question,
			IsAnswerReady: g.IsAnswerReady,
		},
	}

	gameState := QuestionState{
		GameStateID: g.GameStateID,
		Players:     players,
		Round:       int(g.Round),
		RoundType:   g.RoundType,
		Deadline:    time.Until(g.SubmitDeadline.Time),
	}

	return gameState, nil
}
