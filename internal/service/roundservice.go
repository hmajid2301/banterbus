package service

import (
	"context"
	"fmt"
	"time"

	sqlc "gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type RoundService struct {
	store      Storer
	randomizer Randomizer
}

func NewRoundService(store Storer, randomizer Randomizer) *RoundService {
	return &RoundService{store: store, randomizer: randomizer}
}

func (r *RoundService) SubmitAnswer(ctx context.Context, playerID string, answer string, submittedAt time.Time) error {
	room, err := r.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}

	if room.RoomState != sqlc.ROOMSTATE_PLAYING.String() {
		return fmt.Errorf("room is not in PLAYING state")
	}

	round, err := r.store.GetLatestRoundByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}

	if submittedAt.After(round.SubmitDeadline) {
		return fmt.Errorf("answer submission deadline has passed")
	}

	_, err = r.store.AddFibbingItAnswer(ctx, sqlc.AddFibbingItAnswerParams{
		ID:       r.randomizer.GetID(),
		RoundID:  round.ID,
		PlayerID: playerID,
		Answer:   answer,
	})

	return err
}

func (r *RoundService) UpdateStateToVoting(
	ctx context.Context,
	players []PlayerWithRole,
	gameStateID string,
	deadline time.Time,
) ([]VotingPlayer, error) {
	_, err := r.store.UpdateGameState(ctx, sqlc.UpdateGameStateParams{
		ID:             gameStateID,
		SubmitDeadline: deadline,
		State:          sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
	})
	if err != nil {
		return nil, err
	}

	var votingPlayer []VotingPlayer
	for _, player := range players {
		votingPlayer = append(votingPlayer, VotingPlayer{
			ID:       player.ID,
			Nickname: player.Nickname,
			Avatar:   player.Avatar,
		})
	}

	return votingPlayer, nil
}

func (r *RoundService) SubmitVote(
	ctx context.Context,
	playerID string,
	votedNickname string,
	submittedAt time.Time,
) ([]VotingPlayer, error) {
	gameState, err := r.store.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	if gameState.State != sqlc.GAMESTATE_FIBBING_IT_VOTING.String() {
		return nil, fmt.Errorf("game state is not in FIBBING_IT_VOTING state")
	}

	players, err := r.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return nil, err
	}

	votedPlayerID := ""
	for _, p := range players {
		if p.Nickname == votedNickname {
			if p.ID == playerID {
				return nil, fmt.Errorf("cannot vote for yourself")
			}

			votedPlayerID = p.ID
		}
	}

	if votedPlayerID == "" {
		return nil, fmt.Errorf("player with nickname %s not found", votedNickname)
	}

	round, err := r.store.GetLatestRoundByPlayerID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	if submittedAt.After(round.SubmitDeadline) {
		return nil, fmt.Errorf("answer submission deadline has passed")
	}

	votes, err := r.store.CountVotesByRoundID(ctx, round.ID)
	if err != nil {
		return nil, err
	}

	var votingPlayers []VotingPlayer
	for _, p := range votes {
		votingPlayers = append(votingPlayers, VotingPlayer{
			ID:       p.VotedForPlayerID,
			Nickname: p.Nickname,
			Avatar:   p.Avatar,
			Votes:    int(p.VoteCount),
		})
	}

	return votingPlayers, err
}
