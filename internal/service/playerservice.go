package service

import (
	"context"
	"fmt"
	"time"

	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type PlayerService struct {
	store      Storer
	randomizer Randomizer
}

func NewPlayerService(store Storer, randomizer Randomizer) *PlayerService {
	return &PlayerService{store: store, randomizer: randomizer}
}

func (p *PlayerService) UpdateNickname(ctx context.Context, nickname string, playerID string) (Lobby, error) {
	room, err := p.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	if room.RoomState != sqlc.ROOMSTATE_CREATED.String() {
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

	_, err = p.store.UpdateNickname(ctx, sqlc.UpdateNicknameParams{
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

func (p *PlayerService) GenerateNewAvatar(ctx context.Context, playerID string) (Lobby, error) {
	room, err := p.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	if room.RoomState != sqlc.ROOMSTATE_CREATED.String() {
		return Lobby{}, fmt.Errorf("room is not in CREATED state")
	}

	avatar := p.randomizer.GetAvatar()

	_, err = p.store.UpdateAvatar(ctx, sqlc.UpdateAvatarParams{
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

func (p *PlayerService) TogglePlayerIsReady(ctx context.Context, playerID string) (Lobby, error) {
	room, err := p.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	if room.RoomState != sqlc.ROOMSTATE_CREATED.String() {
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
func (p *PlayerService) GetRoomState(ctx context.Context, playerID string) (sqlc.RoomState, error) {
	room, err := p.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return sqlc.ROOMSTATE_CREATED, err
	}

	roomState, err := sqlc.RoomStateFromString(room.RoomState)
	return roomState, err
}

func (p *PlayerService) GetLobby(ctx context.Context, playerID string) (Lobby, error) {
	players, err := p.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return Lobby{}, err
	}

	room := getLobbyPlayers(players, players[0].RoomCode)
	return room, err
}
func (p *PlayerService) GetGameState(ctx context.Context, playerID string) (sqlc.GameStateEnum, error) {
	game, err := p.store.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return sqlc.GAMESTATE_FIBBING_IT_SHOW_QUESTION, err
	}

	gameState, err := sqlc.GameStateFromString(game.State)
	return gameState, err
}

func (p *PlayerService) GetQuestionState(ctx context.Context, playerID string) (QuestionState, error) {
	g, err := p.store.GetCurrentQuestionByPlayerID(ctx, playerID)
	if err != nil {
		return QuestionState{}, err
	}

	question := g.NormalQuestion.String
	if g.FibberQuestion.Valid {
		question = g.FibberQuestion.String
	}

	players := []PlayerWithRole{
		{
			ID:       g.ID,
			Nickname: g.Nickname,
			Role:     g.PlayerRole.String,
			Avatar:   g.Avatar,
			Question: question,
		},
	}

	gameState := QuestionState{
		Players:   players,
		Round:     int(g.Round),
		RoundType: g.RoundType,
		RoomCode:  g.RoomCode,
		Deadline:  time.Until(g.SubmitDeadline),
	}

	return gameState, nil
}
func (p *PlayerService) GetVotingState(ctx context.Context, playerID string) (VotingState, error) {
	round, err := p.store.GetLatestRoundByPlayerID(ctx, playerID)
	if err != nil {
		return VotingState{}, err
	}

	votes, err := p.store.GetVotingState(ctx, round.ID)
	if err != nil {
		return VotingState{}, err
	}

	var votingPlayers []PlayerWithVoting
	for _, p := range votes {
		votingPlayers = append(votingPlayers, PlayerWithVoting{
			ID:       p.VotedForPlayerID,
			Nickname: p.Nickname,
			Avatar:   string(p.Avatar),
			Votes:    int(p.VoteCount),
		})
	}

	if len(votingPlayers) == 0 {
		return VotingState{}, fmt.Errorf("no players found in room")
	}

	votingState := VotingState{
		Round:    int(round.Round),
		Players:  votingPlayers,
		Question: votes[0].Question,
		Deadline: time.Until(votes[0].SubmitDeadline),
	}

	return votingState, nil
}
