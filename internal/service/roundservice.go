package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type RoundService struct {
	store      Storer
	randomizer Randomizer
}

func NewRoundService(store Storer, randomizer Randomizer) *RoundService {
	return &RoundService{store: store, randomizer: randomizer}
}

var ErrMustSubmitAnswer = fmt.Errorf("must submit answer first")

func (r *RoundService) SubmitAnswer(
	ctx context.Context,
	playerID uuid.UUID,
	answer string,
	submittedAt time.Time,
) error {
	room, err := r.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}

	if room.RoomState != db.ROOMSTATE_PLAYING.String() {
		return fmt.Errorf("room is not in PLAYING state")
	}

	round, err := r.store.GetLatestRoundByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}

	if submittedAt.After(round.SubmitDeadline.Time) {
		return fmt.Errorf("answer submission deadline has passed")
	}

	_, err = r.store.AddFibbingItAnswer(ctx, db.AddFibbingItAnswerParams{
		ID:       r.randomizer.GetID(),
		RoundID:  round.ID,
		PlayerID: playerID,
		Answer:   answer,
	})

	return err
}

func (r *RoundService) ToggleAnswerIsReady(
	ctx context.Context,
	playerID uuid.UUID,
	submittedAt time.Time,
) (bool, error) {
	gameState, err := r.store.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return false, err
	}

	if gameState.State != db.GAMESTATE_FIBBING_IT_SHOW_QUESTION.String() {
		return false, fmt.Errorf("room game state is not in FIBBING_IT_SHOW_QUESTION state")
	}

	if submittedAt.After(gameState.SubmitDeadline.Time) {
		return false, fmt.Errorf("toggle ready deadline has passed")
	}

	_, err = r.store.ToggleAnswerIsReady(ctx, playerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrMustSubmitAnswer
		}
		return false, err
	}

	allReady, err := r.store.GetAllPlayerAnswerIsReady(ctx, playerID)
	return allReady, err
}

func (r *RoundService) UpdateStateToVoting(
	ctx context.Context,
	gameStateID uuid.UUID,
	deadline time.Time,
) (VotingState, error) {
	game, err := r.store.GetGameState(ctx, gameStateID)
	if err != nil {
		return VotingState{}, err
	}

	gameState, err := db.GameStateFromString(game.State)
	if err != nil {
		return VotingState{}, err
	} else if gameState != db.GAMESTATE_FIBBING_IT_SHOW_QUESTION {
		return VotingState{}, fmt.Errorf("game state is not in FIBBING_IT_SHOW_QUESTION state")
	}

	_, err = r.store.UpdateGameState(ctx, db.UpdateGameStateParams{
		ID:             gameStateID,
		SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
		State:          db.GAMESTATE_FIBBING_IT_VOTING.String(),
	})
	if err != nil {
		return VotingState{}, err
	}

	round, err := r.store.GetLatestRoundByGameStateID(ctx, gameStateID)
	if err != nil {
		return VotingState{}, err
	}

	votingState, err := r.getVotingState(ctx, round.ID, round.Round)
	return votingState, err
}

func (r *RoundService) SubmitVote(
	ctx context.Context,
	playerID uuid.UUID,
	votedNickname string,
	submittedAt time.Time,
) (VotingState, error) {
	gameState, err := r.store.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return VotingState{}, err
	}

	if gameState.State != db.GAMESTATE_FIBBING_IT_VOTING.String() {
		return VotingState{}, fmt.Errorf("game state is not in FIBBING_IT_VOTING state")
	}

	players, err := r.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return VotingState{}, err
	}

	var votedPlayerID uuid.UUID
	for _, p := range players {
		if p.Nickname == votedNickname {
			if p.ID == playerID {
				return VotingState{}, fmt.Errorf("cannot vote for yourself")
			}

			votedPlayerID = p.ID
		}
	}

	if votedPlayerID == uuid.Nil {
		return VotingState{}, fmt.Errorf("player with nickname %s not found", votedNickname)
	}

	round, err := r.store.GetLatestRoundByPlayerID(ctx, playerID)
	if err != nil {
		return VotingState{}, err
	}

	if submittedAt.After(round.SubmitDeadline.Time) {
		return VotingState{}, fmt.Errorf("answer submission deadline has passed")
	}

	playersWithVoteAndAnswers, err := r.store.GetVotingState(ctx, round.ID)
	if err != nil {
		return VotingState{}, err
	}

	var votingPlayers []PlayerWithVoting
	for _, p := range playersWithVoteAndAnswers {
		voteCount := 0
		if vc, ok := p.Votes.(int); ok {
			voteCount = vc
		}
		votingPlayers = append(votingPlayers, PlayerWithVoting{
			ID:       p.PlayerID,
			Nickname: p.Nickname,
			Avatar:   string(p.Avatar),
			Votes:    voteCount,
		})
	}

	if len(votingPlayers) == 0 {
		return VotingState{}, fmt.Errorf("no players in room")
	}

	player := playersWithVoteAndAnswers[0]
	votingState := VotingState{
		Players:  votingPlayers,
		Question: player.Question,
		Round:    int(player.Round),
		Deadline: time.Until(round.SubmitDeadline.Time),
	}

	return votingState, err
}

func (r *RoundService) GetVotingState(ctx context.Context, playerID uuid.UUID) (VotingState, error) {
	round, err := r.store.GetLatestRoundByPlayerID(ctx, playerID)
	if err != nil {
		return VotingState{}, err
	}

	votingState, err := r.getVotingState(ctx, round.ID, round.Round)
	return votingState, err
}

func (r *RoundService) getVotingState(ctx context.Context, roundID uuid.UUID, round int32) (VotingState, error) {
	votes, err := r.store.GetVotingState(ctx, roundID)
	if err != nil {
		return VotingState{}, err
	}

	var votingPlayers []PlayerWithVoting
	for _, p := range votes {
		voteCount := 0
		if vc, ok := p.Votes.(int); ok {
			voteCount = vc
		}
		votingPlayers = append(votingPlayers, PlayerWithVoting{
			ID:       p.PlayerID,
			Nickname: p.Nickname,
			Avatar:   string(p.Avatar),
			Votes:    voteCount,
		})
	}

	if len(votingPlayers) == 0 {
		return VotingState{}, fmt.Errorf("no players in room")
	}

	votingState := VotingState{
		Round:    int(round),
		Players:  votingPlayers,
		Question: votes[0].Question,
		Deadline: time.Until(votes[0].SubmitDeadline.Time),
	}
	return votingState, nil
}
