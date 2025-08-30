package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

const (
	RoundTypeFreeForm       = "free_form"
	RoundTypeMultipleChoice = "multiple_choice"
	RoundTypeMostLikely     = "most_likely"
	DefaultMaxRounds        = 3
	MaxAnswerLength         = 500 // Maximum characters allowed in an answer
)

type RoundService struct {
	store         Storer
	randomizer    Randomizer
	defaultLocale string
}

func NewRoundService(store Storer, randomizer Randomizer, defaultLocale string) *RoundService {
	return &RoundService{store: store, randomizer: randomizer, defaultLocale: defaultLocale}
}

var ErrMustSubmitAnswer = errors.New("must submit answer first")
var ErrMustSubmitVote = errors.New("must submit vote first")

func (r *RoundService) SubmitAnswer(
	ctx context.Context,
	playerID uuid.UUID,
	answer string,
	submittedAt time.Time,
) error {
	// Validate answer length to prevent memory exhaustion
	if len(answer) > MaxAnswerLength {
		return fmt.Errorf("answer too long: %d characters (max %d)", len(answer), MaxAnswerLength)
	}

	// Validate answer is not empty after trimming
	trimmedAnswer := fmt.Sprintf("%s", answer) // Create copy to avoid modifying original
	if len(trimmedAnswer) == 0 {
		return errors.New("answer cannot be empty")
	}

	room, err := r.store.GetRoomByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}

	// TODO: check game state
	if room.RoomState != db.Playing.String() {
		return errors.New("room is not in PLAYING state")
	}

	round, err := r.store.GetLatestRoundByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}

	// TODO: update logic to match toggle answer is ready
	if submittedAt.After(round.SubmitDeadline.Time) {
		return errors.New("answer submission deadline has passed")
	}

	answers, err := r.getValidAnswers(ctx, round.RoundType, playerID)
	if err != nil {
		return err
	}

	if len(answers) > 0 {
		isAnswerValid := false
		for _, validAnswer := range answers {
			if answer == validAnswer {
				isAnswerValid = true
			}
		}

		if !isAnswerValid {
			msg := fmt.Sprintf("answer received %s must be one of %s", answer, answers)
			return errors.New(msg)
		}
	}

	answerID, err := r.randomizer.GetID()
	if err != nil {
		return err
	}
	_, err = r.store.UpsertFibbingItAnswer(ctx, db.UpsertFibbingItAnswerParams{
		ID:       answerID,
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

	if gameState.State != db.FibbingITQuestion.String() {
		return false, errors.New("room game state is not in FIBBING_IT_QUESTION state")
	}

	if submittedAt.After(gameState.SubmitDeadline.Time) {
		return false, errors.New("toggle ready deadline has passed")
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
	} else if gameState != db.FibbingITQuestion {
		return VotingState{}, errors.New("game state is not in FIBBING_IT_QUESTION state")
	}

	_, err = r.store.UpdateGameState(ctx, db.UpdateGameStateParams{
		ID:             gameStateID,
		SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
		State:          db.FibbingItVoting.String(),
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

	if gameState.State != db.FibbingItVoting.String() {
		return VotingState{}, errors.New("game state is not in FIBBING_IT_VOTING state")
	}

	players, err := r.store.GetAllPlayersInRoom(ctx, playerID)
	if err != nil {
		return VotingState{}, err
	}

	var votedPlayerID uuid.UUID
	for _, p := range players {
		if p.Nickname == votedNickname {
			if p.ID == playerID {
				return VotingState{}, errors.New("cannot vote for yourself")
			}

			votedPlayerID = p.ID
		}
	}

	if votedPlayerID == uuid.Nil {
		return VotingState{}, errors.New(fmt.Sprintf("player with nickname %s not found", votedNickname))
	}

	round, err := r.store.GetLatestRoundByPlayerID(ctx, playerID)
	if err != nil {
		return VotingState{}, err
	}

	if submittedAt.After(round.SubmitDeadline.Time) {
		return VotingState{}, errors.New("answer submission deadline has passed")
	}

	voteID, err := r.randomizer.GetID()
	if err != nil {
		return VotingState{}, err
	}
	err = r.store.UpsertFibbingItVote(ctx, db.UpsertFibbingItVoteParams{
		ID:               voteID,
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
		voteCount := int(p.Votes)
		votingPlayers = append(votingPlayers, PlayerWithVoting{
			ID:       p.PlayerID,
			Nickname: p.Nickname,
			Avatar:   p.Avatar,
			Votes:    voteCount,
			Answer:   p.Answer.String,
		})
	}

	if len(votingPlayers) == 0 {
		return VotingState{}, errors.New("no players in room")
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
		voteCount := int(p.Votes)

		if p.Role.String != FibberRole {
			normalQuestion = p.Question
		}

		votingPlayers = append(votingPlayers, PlayerWithVoting{
			ID:       p.PlayerID,
			Nickname: p.Nickname,
			Avatar:   p.Avatar,
			Votes:    voteCount,
			Answer:   p.Answer.String,
			IsReady:  p.IsReady,
			Role:     p.Role.String,
		})
	}

	sort.Slice(votingPlayers, func(i, j int) bool {
		return votingPlayers[i].Nickname > votingPlayers[j].Nickname
	})

	if len(votingPlayers) == 0 {
		return VotingState{}, errors.New("no players in room")
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

	if gameState.State != db.FibbingItVoting.String() {
		return false, errors.New("room game state is not in FIBBING_IT_VOTING state")
	}

	if submittedAt.After(gameState.SubmitDeadline.Time) {
		return false, errors.New("toggle ready deadline has passed")
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

func (r *RoundService) AreAllPlayersVotingReady(ctx context.Context, gameStateID uuid.UUID) (bool, error) {
	game, err := r.store.GetGameState(ctx, gameStateID)
	if err != nil {
		return false, err
	}

	if game.State != db.FibbingItVoting.String() {
		return false, errors.New("game state is not in FIBBING_IT_VOTING state")
	}

	// Get any player from this room to use for the voting status check
	players, err := r.store.GetAllPlayersByGameStateID(ctx, gameStateID)
	if err != nil {
		return false, err
	}

	if len(players) == 0 {
		return false, errors.New("no players in room")
	}

	// Use the first player to check if all players are ready for voting
	allReady, err := r.store.GetAllPlayersVotingIsReady(ctx, players[0].ID)
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
	} else if gameState != db.FibbingItVoting {
		return RevealRoleState{}, errors.New("game state is not in FIBBING_IT_VOTING state")
	}

	_, err = r.store.UpdateGameState(ctx, db.UpdateGameStateParams{
		ID:             gameStateID,
		SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
		State:          db.FibbingItRevealRole.String(),
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

	reveal := RevealRoleState{
		Deadline:     time.Until(deadline),
		Round:        votingState.Round,
		RoundType:    round.RoundType,
		ShouldReveal: false,
	}
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
	nextRound bool,
) (QuestionState, error) {
	game, err := r.store.GetGameState(ctx, gameStateID)
	if err != nil {
		return QuestionState{}, err
	}

	gameState, err := db.GameStateFromString(game.State)
	if err != nil {
		return QuestionState{}, err
	} else if gameState != db.FibbingItRevealRole && gameState != db.FibbingItScoring {
		return QuestionState{}, errors.New("game state is not in FIBBING_IT_REVEAL_ROLE state or FIBBING_IT_SCORING state")
	}

	_, err = r.store.UpdateGameState(ctx, db.UpdateGameStateParams{
		ID:             gameStateID,
		SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
		State:          db.FibbingITQuestion.String(),
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
	fibberLoc := -1

	// INFO: If we next round means that we need to find the next fibber
	// Else it should be the same fibber player used, so find the current fibber in the player slice.
	var maxRounds int32 = DefaultMaxRounds
	if roundNumber == maxRounds+1 || nextRound {
		nextRoundType := getNextRoundType(roundType)
		if nextRoundType == "" {
			// No more round types, this means we've completed all game types
			return QuestionState{}, errors.New("game completed - no more round types available")
		}
		roundType = nextRoundType
		roundNumber = 1
		fibberLoc = r.randomizer.GetFibberIndex(len(players))
	} else {
		fibber, err := r.store.GetFibberByRoundID(ctx, round.ID)
		if err != nil {
			return QuestionState{}, fmt.Errorf("failed to get fibber in round: %w", err)
		}

		for i, player := range players {
			if player.ID == fibber.PlayerID {
				fibberLoc = i
			}
		}
	}

	normalsQuestions, fibberQuestions, err := getQuestions(ctx, r.store, "fibbing_it", roundType)
	if err != nil {
		return QuestionState{}, errors.New(err.Error())
	}

	if len(normalsQuestions) == 0 {
		return QuestionState{}, errors.New("no normal questions available")
	}

	if len(fibberQuestions) == 0 {
		return QuestionState{}, errors.New("no fibber questions available")
	}

	if fibberLoc == -1 {
		return QuestionState{}, errors.New("failed to set fibber location in players slice")
	}

	newRound := db.NewRoundArgs{
		GameStateID:       gameStateID,
		NormalsQuestionID: normalsQuestions[0].QuestionID,
		FibberQuestionID:  fibberQuestions[0].QuestionID,
		RoundType:         roundType,
		Round:             roundNumber,
		Players:           players,
		FibberLoc:         fibberLoc,
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

		if i == fibberLoc {
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
		if roundType == RoundTypeMultipleChoice {
			// TODO: localise use player locale?
			answers = []string{"Strongly Agree", "Agree", "Neutral", "Disagree", "Strongly Disagree"}
		} else if roundType == RoundTypeMostLikely {
			for _, p := range players {
				answers = append(answers, p.Nickname)
			}
			slices.Sort(answers)
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

	answers, err := r.getValidAnswers(ctx, g.RoundType, playerID)
	if err != nil {
		return QuestionState{}, err
	}

	players := []PlayerWithRole{
		{
			ID:              g.PlayerID,
			Role:            g.Role.String,
			Question:        g.Question.String,
			IsAnswerReady:   g.IsAnswerReady,
			PossibleAnswers: answers,
			CurrentAnswer:   g.CurrentAnswer,
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
	} else if gameState != db.FibbingItRevealRole {
		return ScoreState{}, errors.New("game state is not in FIBBING_IT_REVEAL_ROLE state")
	}

	_, err = r.store.UpdateGameState(ctx, db.UpdateGameStateParams{
		ID:             gameStateID,
		SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
		State:          db.FibbingItScoring.String(),
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

func (r *RoundService) GetGameState(ctx context.Context, playerID uuid.UUID) (db.FibbingItGameState, error) {
	game, err := r.store.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return db.FibbingITQuestion, err
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

	scoredByPlayerID, err := r.store.GetTotalScoresByGameStateID(ctx, db.GetTotalScoresByGameStateIDParams{
		ID:   gameStateID,
		ID_2: round.ID,
	})
	if err != nil {
		return ScoreState{}, nil, err
	}

	currentScoreMap := map[uuid.UUID]int{}
	for _, p := range scoredByPlayerID {
		currentScoreMap[p.PlayerID] = int(p.TotalScore)
	}

	// INFO: This shouldn't happen.
	if len(allVotesInRoundType) == 0 {
		return ScoreState{}, nil, errors.New("no players in game")
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

		// INFO: If user has an existing score add it on so we can show their total score.
		// If this is the first round of scoring, then there won't be a score to show.
		if currentScore, ok := currentScoreMap[p.VoterID]; ok {
			player.Score += currentScore
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

func (r *RoundService) UpdateStateToWinner(
	ctx context.Context,
	gameStateID uuid.UUID,
	deadline time.Time,
) (WinnerState, error) {
	game, err := r.store.GetGameState(ctx, gameStateID)
	if err != nil {
		return WinnerState{}, err
	}

	gameState, err := db.GameStateFromString(game.State)
	if err != nil {
		return WinnerState{}, err
	} else if gameState != db.FibbingItScoring {
		return WinnerState{}, errors.New("game state is not in FIBBING_IT_SCORING_STATE state")
	}

	_, err = r.store.UpdateGameState(ctx, db.UpdateGameStateParams{
		ID:             gameStateID,
		SubmitDeadline: pgtype.Timestamp{Time: deadline, Valid: true},
		State:          db.FibbingItWinner.String(),
	})
	if err != nil {
		return WinnerState{}, err
	}

	return r.getWinnerState(ctx, gameStateID)
}

func (r *RoundService) GetWinnerState(ctx context.Context, playerID uuid.UUID) (WinnerState, error) {
	gameState, err := r.store.GetGameStateByPlayerID(ctx, playerID)
	if err != nil {
		return WinnerState{}, err
	}

	return r.getWinnerState(ctx, gameState.ID)
}

func (r *RoundService) getWinnerState(ctx context.Context, gameStateID uuid.UUID) (WinnerState, error) {
	// INFO: Query adds all scored that don't include a certain round ID, so we use a fake round ID, so it adds
	// all player scores.
	fakeRoundID, err := r.randomizer.GetID()
	if err != nil {
		return WinnerState{}, err
	}
	scoredByPlayerID, err := r.store.GetTotalScoresByGameStateID(ctx, db.GetTotalScoresByGameStateIDParams{
		ID:   gameStateID,
		ID_2: fakeRoundID,
	})

	sort.Slice(scoredByPlayerID, func(i, j int) bool {
		return scoredByPlayerID[i].TotalScore > scoredByPlayerID[j].TotalScore
	})
	if err != nil {
		return WinnerState{}, err
	}

	players := []PlayerWithScoring{}
	for _, p := range scoredByPlayerID {
		player := PlayerWithScoring{
			ID:       p.PlayerID,
			Score:    int(p.TotalScore),
			Avatar:   p.Avatar,
			Nickname: p.Nickname,
		}

		players = append(players, player)
	}

	return WinnerState{
		Players: players,
	}, nil
}

func getNextRoundType(roundType string) string {
	nextRoundMap := map[string]string{
		RoundTypeFreeForm:       RoundTypeMultipleChoice,
		RoundTypeMultipleChoice: RoundTypeMostLikely,
		RoundTypeMostLikely:     "", // No next round type - game should end
	}

	return nextRoundMap[roundType]
}

func (r RoundService) getValidAnswers(ctx context.Context, roundType string, playerID uuid.UUID) ([]string, error) {
	answers := []string{}
	if roundType == RoundTypeMultipleChoice {
		answers = []string{"Strongly Agree", "Agree", "Neutral", "Disagree", "Strongly Disagree"}
	} else if roundType == RoundTypeMostLikely {
		players, err := r.store.GetAllPlayersInRoom(ctx, playerID)
		if err != nil {
			return nil, err
		}

		for _, player := range players {
			answers = append(answers, player.Nickname)
		}
	}
	return answers, nil
}

func (r *RoundService) FinishGame(ctx context.Context, gameStateID uuid.UUID) error {
	game, err := r.store.GetGameState(ctx, gameStateID)
	if err != nil {
		return err
	}

	// TODO: check current room state

	_, err = r.store.UpdateRoomState(ctx, db.UpdateRoomStateParams{RoomState: db.Finished.String(), ID: game.RoomID})
	if err != nil {
		return err
	}

	return nil
}
