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
	updateVoting UpdateVotingState,
) (VotingState, error) {
	_, err := r.store.UpdateGameState(ctx, sqlc.UpdateGameStateParams{
		ID:             updateVoting.GameStateID,
		SubmitDeadline: updateVoting.Deadline,
		State:          sqlc.GAMESTATE_FIBBING_IT_VOTING.String(),
	})
	if err != nil {
		return VotingState{}, err
	}

	var votingPlayer []PlayerWithVoting
	var normalQuestion string
	for _, player := range updateVoting.Players {
		if player.Role == "normal" {
			normalQuestion = player.Question
		}
		votingPlayer = append(votingPlayer, PlayerWithVoting{
			ID:       player.ID,
			Nickname: player.Nickname,
			Avatar:   string(player.Avatar),
		})
	}

	// INFO: Shouldn't happen
	if len(votingPlayer) == 0 {
		return VotingState{}, fmt.Errorf("no players in room")
	}

	// TODO: handle localistion will need to return one per player i think
	votingState := VotingState{
		Question: normalQuestion,
		Players:  votingPlayer,
		Round:    updateVoting.Round,
	}

	return votingState, nil
}

func (r *RoundService) SubmitVote(
	ctx context.Context,
	playerID string,
	votedNickname string,
	submittedAt time.Time,
) (VotingState, error) {
	gameState, err := r.store.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return VotingState{}, err
	}

	if gameState.State != sqlc.GAMESTATE_FIBBING_IT_VOTING.String() {
		return VotingState{}, fmt.Errorf("game state is not in FIBBING_IT_VOTING state")
	}

	players, err := r.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return VotingState{}, err
	}

	votedPlayerID := ""
	for _, p := range players {
		if p.Nickname == votedNickname {
			if p.ID == playerID {
				return VotingState{}, fmt.Errorf("cannot vote for yourself")
			}

			votedPlayerID = p.ID
		}
	}

	if votedPlayerID == "" {
		return VotingState{}, fmt.Errorf("player with nickname %s not found", votedNickname)
	}

	round, err := r.store.GetLatestRoundByPlayerID(ctx, playerID)
	if err != nil {
		return VotingState{}, err
	}

	if submittedAt.After(round.SubmitDeadline) {
		return VotingState{}, fmt.Errorf("answer submission deadline has passed")
	}

	playersWithVoteAndAnswers, err := r.store.GetVotingState(ctx, round.ID)
	if err != nil {
		return VotingState{}, err
	}

	var votingPlayers []PlayerWithVoting
	for _, p := range playersWithVoteAndAnswers {
		votingPlayers = append(votingPlayers, PlayerWithVoting{
			ID:       p.VotedForPlayerID,
			Nickname: p.Nickname,
			Avatar:   string(p.Avatar),
			Votes:    int(p.VoteCount),
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
		Deadline: time.Until(round.SubmitDeadline),
	}

	return votingState, err
}
