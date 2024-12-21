package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type RoundService struct {
	store         Storer
	randomizer    Randomizer
	defaultLocale string
}

func NewRoundService(store Storer, randomizer Randomizer, defaultLocale string) *RoundService {
	return &RoundService{store: store, randomizer: randomizer, defaultLocale: defaultLocale}
}

var ErrMustSubmitAnswer = fmt.Errorf("must submit answer first")
var ErrMustSubmitVote = fmt.Errorf("must submit vote first")

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

	// TODO: check game state
	if room.RoomState != db.ROOMSTATE_PLAYING.String() {
		return fmt.Errorf("room is not in PLAYING state")
	}

	round, err := r.store.GetLatestRoundByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}

	// TODO: update logic to match toggle answer is ready
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

	// TODO: no answer submitted should throw an error
	if gameState.State != db.GAMESTATE_FIBBING_IT_QUESTION.String() {
		return false, fmt.Errorf("room game state is not in FIBBING_IT_QUESTION state")
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
	} else if gameState != db.GAMESTATE_FIBBING_IT_QUESTION {
		return VotingState{}, fmt.Errorf("game state is not in FIBBING_IT_QUESTION state")
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

	u := r.randomizer.GetID()
	err = r.store.UpsertFibbingItVote(ctx, db.UpsertFibbingItVoteParams{
		ID:               u,
		RoundID:          round.ID,
		PlayerID:         playerID,
		VotedForPlayerID: votedPlayerID,
	})
	if err != nil {
		return VotingState{}, err
	}

	playersWithVoteAndAnswers, err := r.store.GetVotingState(ctx, round.ID)
	if err != nil {
		return VotingState{}, err
	}

	var votingPlayers []PlayerWithVoting
	for _, p := range playersWithVoteAndAnswers {
		voteCount := 0
		if vc, ok := p.Votes.(int64); ok {
			voteCount = int(vc)
		}
		votingPlayers = append(votingPlayers, PlayerWithVoting{
			ID:       p.PlayerID,
			Nickname: p.Nickname,
			Avatar:   p.Avatar,
			Votes:    voteCount,
			Answer:   p.Answer.String,
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

	var normalQuestion string
	var votingPlayers []PlayerWithVoting
	for _, p := range votes {
		voteCount := 0
		if vc, ok := p.Votes.(int); ok {
			voteCount = vc
		}

		if p.Role.String != FibberRole {
			normalQuestion = p.Question
		}

		votingPlayers = append(votingPlayers, PlayerWithVoting{
			ID:       p.PlayerID,
			Nickname: p.Nickname,
			Avatar:   p.Avatar,
			Votes:    voteCount,
			Answer:   p.Answer.String,
			IsReady:  p.IsReady.Bool,
			Role:     p.Role.String,
		})
	}

	if len(votingPlayers) == 0 {
		return VotingState{}, fmt.Errorf("no players in room")
	}

	votingState := VotingState{
		GameStateID: votes[0].GameStateID,
		Round:       int(round),
		Players:     votingPlayers,
		Question:    normalQuestion,
		Deadline:    time.Until(votes[0].SubmitDeadline.Time),
	}
	return votingState, nil
}

func (r *RoundService) ToggleVotingIsReady(
	ctx context.Context,
	playerID uuid.UUID,
	submittedAt time.Time,
) (bool, error) {
	gameState, err := r.store.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return false, err
	}

	if gameState.State != db.GAMESTATE_FIBBING_IT_VOTING.String() {
		return false, fmt.Errorf("room game state is not in FIBBING_IT_VOTING state")
	}

	if submittedAt.After(gameState.SubmitDeadline.Time) {
		return false, fmt.Errorf("toggle ready deadline has passed")
	}

	_, err = r.store.ToggleVotingIsReady(ctx, playerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrMustSubmitVote
		}
		return false, err
	}

	allReady, err := r.store.GetAllPlayersVotingIsReady(ctx, playerID)
	return allReady, err
}

func (r *RoundService) UpdateStateToReveal(
	ctx context.Context,
	gameStateID uuid.UUID,
	deadline time.Time,
) (RevealRoleState, error) {
	game, err := r.store.GetGameState(ctx, gameStateID)
	if err != nil {
		return RevealRoleState{}, err
	}

	gameState, err := db.GameStateFromString(game.State)
	if err != nil {
		return RevealRoleState{}, err
	} else if gameState != db.GAMESTATE_FIBBING_IT_VOTING {
		return RevealRoleState{}, fmt.Errorf("game state is not in FIBBING_IT_VOTING state")
	}

	_, err = r.store.UpdateGameState(ctx, db.UpdateGameStateParams{
		ID:             gameStateID,
		SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
		State:          db.GAMESTATE_FIBBING_IT_REVEAL_ROLE.String(),
	})
	if err != nil {
		return RevealRoleState{}, err
	}

	reveal, err := r.getRevealState(ctx, gameStateID, deadline)
	return reveal, err
}

func (r *RoundService) GetRevealState(ctx context.Context, playerID uuid.UUID) (RevealRoleState, error) {
	gameState, err := r.store.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return RevealRoleState{}, err
	}

	revealState, err := r.getRevealState(ctx, gameState.ID, gameState.SubmitDeadline.Time)
	return revealState, err
}

func (r *RoundService) getRevealState(
	ctx context.Context,
	gameStateID uuid.UUID,
	deadline time.Time,
) (RevealRoleState, error) {
	round, err := r.store.GetLatestRoundByGameStateID(ctx, gameStateID)
	if err != nil {
		return RevealRoleState{}, err
	}

	votingState, err := r.getVotingState(ctx, round.ID, round.Round)
	if err != nil {
		return RevealRoleState{}, err
	}

	reveal := RevealRoleState{Deadline: time.Until(deadline), Round: votingState.Round, ShouldReveal: false}
	playerIDs := []uuid.UUID{}
	playersLen := len(votingState.Players)

	for _, p := range votingState.Players {
		playerIDs = append(playerIDs, p.ID)
		if p.Votes == playersLen-1 {
			reveal.VotedForPlayerNickname = p.Nickname
			reveal.VotedForPlayerAvatar = p.Avatar
			reveal.VotedForPlayerRole = p.Role
			reveal.ShouldReveal = true
		}
	}

	reveal.PlayerIDs = playerIDs
	return reveal, nil
}

// TODO: see if we can use this in start game lobbyservice
func (r *RoundService) UpdateStateToQuestion(
	ctx context.Context,
	gameStateID uuid.UUID,
	deadline time.Time,
) (QuestionState, error) {
	game, err := r.store.GetGameState(ctx, gameStateID)
	if err != nil {
		return QuestionState{}, err
	}

	gameState, err := db.GameStateFromString(game.State)
	if err != nil {
		return QuestionState{}, err
	} else if gameState != db.GAMESTATE_FIBBING_IT_REVEAL_ROLE && gameState != db.GAMESTATE_FIBBING_IT_SCORING {
		return QuestionState{}, fmt.Errorf("game state is not in FIBBING_IT_REVEAL_ROLE state or FIBBING_IT_SCORING state")
	}

	_, err = r.store.UpdateGameState(ctx, db.UpdateGameStateParams{
		ID:             gameStateID,
		SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
		State:          db.GAMESTATE_FIBBING_IT_QUESTION.String(),
	})
	if err != nil {
		return QuestionState{}, err
	}

	players, err := r.store.GetAllPlayersByGameStateID(ctx, gameStateID)
	if err != nil {
		return QuestionState{}, err
	}

	round, err := r.store.GetLatestRoundByGameStateID(ctx, gameStateID)
	if err != nil {
		return QuestionState{}, err
	}

	roundType := round.RoundType
	roundNumber := round.Round + 1
	// TODO: move to config
	var maxRounds int32 = 3
	if roundNumber == maxRounds+1 {
		roundType = getNextRoundType(roundType)
		roundNumber = 1
	}

	normalsQuestions, err := r.store.GetRandomQuestionByRound(ctx, db.GetRandomQuestionByRoundParams{
		GameName:  "fibbing_it",
		RoundType: roundType,
	})
	if err != nil {
		return QuestionState{}, err
	}

	fibberQuestions, err := r.store.GetRandomQuestionInGroup(ctx, db.GetRandomQuestionInGroupParams{
		GroupID: normalsQuestions[0].GroupID,
		ID:      normalsQuestions[0].ID,
	})
	if err != nil {
		return QuestionState{}, err
	}

	randomFibberLoc := r.randomizer.GetFibberIndex(len(players))
	newRound := db.NewRoundArgs{
		GameStateID:       gameStateID,
		NormalsQuestionID: normalsQuestions[0].QuestionID,
		FibberQuestionID:  fibberQuestions[0].QuestionID,
		RoundType:         roundType,
		Round:             roundNumber,
		Players:           players,
		FibberLoc:         randomFibberLoc,
	}

	err = r.store.NewRound(ctx, newRound)
	if err != nil {
		return QuestionState{}, err
	}

	playersWithRole := []PlayerWithRole{}
	for i, player := range players {
		role := NormalRole

		var question string
		for _, localeQuestion := range normalsQuestions {
			if localeQuestion.Locale == player.Locale.String {
				question = localeQuestion.Question
			} else if question == "" && localeQuestion.Locale == r.defaultLocale {
				question = localeQuestion.Question
			}
		}

		if i == randomFibberLoc {
			role = FibberRole
			question = ""
			for _, localeQuestion := range fibberQuestions {
				if localeQuestion.Locale == player.Locale.String {
					question = localeQuestion.Question
				} else if question == "" && localeQuestion.Locale == r.defaultLocale {
					question = localeQuestion.Question
				}
			}
		}
		answers := []string{}
		if roundType == "multiple_choice" {
			// TODO: localise use player locale?
			answers = []string{"Strongly Agree", "Agree", "Neutral", "Disagree", "Strongly Disagree"}
		} else if roundType == "most_likely" {
			for _, p := range players {
				if p.ID != player.ID {
					answers = append(answers, p.Nickname)
				}
			}
		}

		playersWithRole = append(playersWithRole, PlayerWithRole{
			ID:              player.ID,
			Role:            role,
			Question:        question,
			IsAnswerReady:   false,
			PossibleAnswers: answers,
		})
	}

	timeLeft := time.Until(deadline)

	questionState := QuestionState{
		GameStateID: gameStateID,
		Players:     playersWithRole,
		Round:       int(roundNumber),
		RoundType:   roundType,
		Deadline:    timeLeft,
	}
	return questionState, nil
}

func (r *RoundService) GetQuestionState(ctx context.Context, playerID uuid.UUID) (QuestionState, error) {
	g, err := r.store.GetCurrentQuestionByPlayerID(ctx, playerID)
	if err != nil {
		return QuestionState{}, err
	}

	answers := []string{}
	if g.RoundType == "multiple_choice" {
		answers = []string{"Strongly Agree", "Agree", "Neutral", "Disagree", "Strongly Disagree"}
	} else if g.RoundType == "most_likely" {
		players, err := r.store.GetAllPlayersInRoom(ctx, playerID)
		if err != nil {
			return QuestionState{}, err
		}
		for _, player := range players {
			if player.ID != playerID {
				answers = append(answers, player.Nickname)
			}
		}
	}

	players := []PlayerWithRole{
		{
			ID:              g.PlayerID,
			Role:            g.Role.String,
			Question:        g.Question.String,
			IsAnswerReady:   g.IsAnswerReady,
			PossibleAnswers: answers,
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

func (r *RoundService) UpdateStateToScore(
	ctx context.Context,
	gameStateID uuid.UUID,
	deadline time.Time,
	scoring Scoring,
) (ScoreState, error) {
	game, err := r.store.GetGameState(ctx, gameStateID)
	if err != nil {
		return ScoreState{}, err
	}

	gameState, err := db.GameStateFromString(game.State)
	if err != nil {
		return ScoreState{}, err
	} else if gameState != db.GAMESTATE_FIBBING_IT_REVEAL_ROLE {
		return ScoreState{}, fmt.Errorf("game state is not in FIBBING_IT_REVEAL_ROLE state")
	}

	_, err = r.store.UpdateGameState(ctx, db.UpdateGameStateParams{
		ID:             gameStateID,
		SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
		State:          db.GAMESTATE_FIBBING_IT_SCORING.String(),
	})
	if err != nil {
		return ScoreState{}, err
	}

	scoringState, dbPlayerScores, err := r.getScoreState(ctx, scoring, gameStateID, deadline)
	if err != nil {
		return ScoreState{}, err
	}

	err = r.store.NewScores(ctx, db.NewScoresArgs{Players: dbPlayerScores})
	if err != nil {
		return ScoreState{}, err
	}

	return scoringState, nil
}

func (r *RoundService) GetGameState(ctx context.Context, playerID uuid.UUID) (db.GameStateEnum, error) {
	game, err := r.store.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return db.GAMESTATE_FIBBING_IT_QUESTION, err
	}

	gameState, err := db.GameStateFromString(game.State)
	return gameState, err
}

func (r *RoundService) GetScoreState(ctx context.Context, scoring Scoring, playerID uuid.UUID) (ScoreState, error) {
	gameState, err := r.store.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return ScoreState{}, err
	}

	scoreState, _, err := r.getScoreState(ctx, scoring, gameState.ID, gameState.SubmitDeadline.Time)
	return scoreState, err
}

func (r *RoundService) getScoreState(
	ctx context.Context,
	scoring Scoring,
	gameStateID uuid.UUID,
	deadline time.Time,
) (ScoreState, []db.AddFibbingItScoreParams, error) {
	playerScoreMap := map[uuid.UUID]PlayerWithScoring{}
	allVotesInRoundType, err := r.store.GetAllVotesForRoundByGameStateID(ctx, gameStateID)
	if err != nil {
		return ScoreState{}, nil, err
	}

	allPlayers, err := r.store.GetAllPlayersByGameStateID(ctx, gameStateID)
	if err != nil {
		return ScoreState{}, nil, err
	}
	round, err := r.store.GetLatestRoundByGameStateID(ctx, gameStateID)
	if err != nil {
		return ScoreState{}, nil, err
	}

	// INFO: This shouldn't happen.
	if len(allVotesInRoundType) == 0 {
		return ScoreState{}, nil, fmt.Errorf("no players in game")
	}

	fibberVotesThisRound := 0
	fibberCaught := false

	// TODO: add score to previous rounds
	fibberID := allVotesInRoundType[0].FibberID
	for _, p := range allVotesInRoundType {
		player, ok := playerScoreMap[p.VoterID]
		if !ok {
			player = PlayerWithScoring{
				ID:       p.VoterID,
				Avatar:   p.VoterAvatar,
				Nickname: p.VoterNickname,
				Score:    0,
			}
		}

		playerScoreMap[p.VoterID] = player
		if p.VotedForID == fibberID {
			player.Score += scoring.GuessedFibber

			if p.RoundID == round.ID {
				fibberVotesThisRound++
			}
		} else if p.VoterID != fibberID {
			player.Score += scoring.FibberEvadeCapture
		}

		playerScoreMap[p.VoterID] = player
	}

	if len(allVotesInRoundType) == fibberVotesThisRound {
		fibberCaught = true
	}

	// INFO: Some players may not have voted, so we want to give them a score of 0. Unless they are fibber,
	// then they get a score for every round they evaded capture.
	playersScore := []PlayerWithScoring{}
	dbPlayerScores := []db.AddFibbingItScoreParams{}
	for _, p := range allPlayers {
		if player, ok := playerScoreMap[p.ID]; ok {
			playersScore = append(playersScore, player)
		} else {
			score := PlayerWithScoring{
				ID:       p.ID,
				Avatar:   p.Avatar,
				Nickname: p.Nickname,
				Score:    0,
			}

			if p.ID == fibberID {
				// INFO: Fibber got caught this round
				roundNumber := round.Round
				if fibberCaught {
					roundNumber--
				}

				score.Score = scoring.FibberEvadeCapture * int(roundNumber)
			}
			playersScore = append(playersScore, score)
		}

		dbPlayerScores = append(dbPlayerScores, db.AddFibbingItScoreParams{
			RoundID:  round.ID,
			PlayerID: p.ID,
			//nolint:gosec // disable G115
			Score: int32(playerScoreMap[p.ID].Score),
		})
	}

	sort.Slice(playersScore, func(i, j int) bool {
		return playersScore[i].Score > playersScore[j].Score
	})

	timeLeft := time.Until(deadline)
	scoringState := ScoreState{
		Players:     playersScore,
		Deadline:    timeLeft,
		RoundType:   round.RoundType,
		RoundNumber: int(round.Round),
	}

	return scoringState, dbPlayerScores, nil
}

func getNextRoundType(roundType string) string {
	nextRoundMap := map[string]string{
		"free_form":       "multiple_choice",
		"multiple_choice": "most_likely",
	}

	return nextRoundMap[roundType]
}
