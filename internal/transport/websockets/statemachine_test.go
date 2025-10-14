package websockets_test

import (
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/invopop/ctxi18n"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets"
	mockService "gitlab.com/hmajid2301/banterbus/internal/transport/websockets/mocks"
	"gitlab.com/hmajid2301/banterbus/internal/views"
)

func TestStateMachine(t *testing.T) {
	t.Parallel()

	// Initialize i18n for all tests
	ctxi18n.LoadWithDefault(views.Locales, "en-GB")

	t.Run("Should successfully complete question state and move to voting", func(t *testing.T) {
		t.Parallel()

		roundService := mockService.NewMockRoundServicer(t)
		lobbyService := mockService.NewMockLobbyServicer(t)
		playerService := mockService.NewMockPlayerServicer(t)
		log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
		websocketer := mockService.NewMockWebsocketer(t)

		ctx := t.Context()
		conf, err := config.LoadConfig(ctx)
		require.NoError(t, err)

		rules, err := views.RuleMarkdown("fibbing_it")
		require.NoError(t, err)

		conf.Timings.ShowQuestionScreenFor = time.Millisecond * 1
		conf.Timings.ShowVotingScreenFor = time.Millisecond * 1
		conf.Timings.ShowRevealScreenFor = time.Millisecond * 1
		conf.Timings.ShowScoreScreenFor = time.Millisecond * 1
		conf.Timings.ShowWinnerScreenFor = time.Millisecond * 1

		sub := websockets.NewSubscriber(lobbyService, playerService, roundService, log, websocketer, conf, rules)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		q := websockets.QuestionState{
			GameStateID: u,
			Subscriber:  *sub,
			NextRound:   false,
		}

		p1, err := uuid.NewV7()
		require.NoError(t, err)
		p2, err := uuid.NewV7()
		require.NoError(t, err)

		// Mock GetGameState call that happens at the start of the state machine
		roundService.EXPECT().
			GetGameState(mock.Anything, u).
			Return(db.FibbingITQuestion, nil)

		roundService.EXPECT().
			UpdateStateToQuestion(mock.Anything, u, mock.AnythingOfType("time.Time"), false).
			Return(service.QuestionState{
				GameStateID: u,
				Players: []service.PlayerWithRole{
					{
						ID:       p1,
						Role:     "fibber",
						Question: "Fibber question",
					},
					{
						ID:       p2,
						Role:     "normal",
						Question: "Normal question",
					},
				},
				Round:     1,
				RoundType: "free_form",
				Deadline:  time.Second * 1,
			}, nil)

		playerService.EXPECT().GetPlayerByID(mock.Anything, p1).Return(db.Player{Locale: pgtype.Text{String: "en-US", Valid: true}}, nil)
		playerService.EXPECT().GetPlayerByID(mock.Anything, p2).Return(db.Player{Locale: pgtype.Text{String: "en-US", Valid: true}}, nil)

		websocketer.EXPECT().Publish(mock.Anything, p1, mock.AnythingOfType("[]uint8")).Return(nil)
		websocketer.EXPECT().Publish(mock.Anything, p2, mock.AnythingOfType("[]uint8")).Return(nil)

		roundService.EXPECT().
			AreAllPlayersAnswerReady(mock.Anything, u).
			Return(false, nil).
			Maybe()

		roundService.EXPECT().
			UpdateStateToVoting(mock.Anything, u, mock.AnythingOfType("time.Time")).
			Return(service.VotingState{}, fmt.Errorf("stop state machine here"))

		q.Start(ctx)
		// INFO: Wait for the voting state goroutine to spin up
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("Should fail to complete question state, because fail to update state to question", func(t *testing.T) {
		t.Parallel()

		roundService := mockService.NewMockRoundServicer(t)
		lobbyService := mockService.NewMockLobbyServicer(t)
		playerService := mockService.NewMockPlayerServicer(t)
		log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
		websocketer := mockService.NewMockWebsocketer(t)

		ctx := t.Context()
		conf, err := config.LoadConfig(ctx)
		require.NoError(t, err)

		rules, err := views.RuleMarkdown("fibbing_it")
		require.NoError(t, err)

		conf.Timings.ShowQuestionScreenFor = time.Millisecond * 1
		conf.Timings.ShowVotingScreenFor = time.Millisecond * 1
		conf.Timings.ShowRevealScreenFor = time.Millisecond * 1
		conf.Timings.ShowScoreScreenFor = time.Millisecond * 1
		conf.Timings.ShowWinnerScreenFor = time.Millisecond * 1

		sub := websockets.NewSubscriber(lobbyService, playerService, roundService, log, websocketer, conf, rules)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		q := websockets.QuestionState{
			GameStateID: u,
			Subscriber:  *sub,
			NextRound:   false,
		}

		roundService.EXPECT().
			UpdateStateToQuestion(mock.Anything, u, mock.AnythingOfType("time.Time"), false).
			Return(service.QuestionState{}, fmt.Errorf("failed to update state"))

		q.Start(ctx)
	})

	t.Run("Should successfully complete voting state and move to reveal", func(t *testing.T) {
		t.Parallel()

		roundService := mockService.NewMockRoundServicer(t)
		lobbyService := mockService.NewMockLobbyServicer(t)
		playerService := mockService.NewMockPlayerServicer(t)
		log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
		websocketer := mockService.NewMockWebsocketer(t)

		ctx := t.Context()
		conf, err := config.LoadConfig(ctx)
		require.NoError(t, err)

		rules, err := views.RuleMarkdown("fibbing_it")
		require.NoError(t, err)

		conf.Timings.ShowQuestionScreenFor = time.Millisecond * 1
		conf.Timings.ShowVotingScreenFor = time.Millisecond * 1
		conf.Timings.ShowRevealScreenFor = time.Millisecond * 1
		conf.Timings.ShowScoreScreenFor = time.Millisecond * 1
		conf.Timings.ShowWinnerScreenFor = time.Millisecond * 1

		sub := websockets.NewSubscriber(lobbyService, playerService, roundService, log, websocketer, conf, rules)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		q := websockets.VotingState{
			GameStateID: u,
			Subscriber:  *sub,
		}

		p1, err := uuid.NewV7()
		require.NoError(t, err)
		p2, err := uuid.NewV7()
		require.NoError(t, err)
		roundService.EXPECT().
			UpdateStateToVoting(mock.Anything, u, mock.AnythingOfType("time.Time")).
			Return(service.VotingState{
				GameStateID: u,
				Players: []service.PlayerWithVoting{
					{
						ID:       p1,
						Role:     "fibber",
						Nickname: "Player1",
						Avatar:   "avatar1.png",
					},
					{
						ID:       p2,
						Role:     "normal",
						Nickname: "Player2",
						Avatar:   "avatar2.png",
					},
				},
				Question: "Who is most likely to...?",
				Round:    1,
				Deadline: time.Second * 1,
			}, nil)

		playerService.EXPECT().GetPlayerByID(mock.Anything, p1).Return(db.Player{Locale: pgtype.Text{String: "en-US", Valid: true}}, nil)
		playerService.EXPECT().GetPlayerByID(mock.Anything, p2).Return(db.Player{Locale: pgtype.Text{String: "en-US", Valid: true}}, nil)

		websocketer.EXPECT().Publish(mock.Anything, p1, mock.AnythingOfType("[]uint8")).Return(nil)
		websocketer.EXPECT().Publish(mock.Anything, p2, mock.AnythingOfType("[]uint8")).Return(nil)

		roundService.EXPECT().
			UpdateStateToReveal(mock.Anything, u, mock.AnythingOfType("time.Time")).
			Return(service.RevealRoleState{}, fmt.Errorf("stop state machine here"))

		q.Start(ctx)
		// INFO: Wait for the reveal state goroutine to spin up
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("Should fail to complete voting state, because fail to update state to voting", func(t *testing.T) {
		t.Parallel()

		roundService := mockService.NewMockRoundServicer(t)
		lobbyService := mockService.NewMockLobbyServicer(t)
		playerService := mockService.NewMockPlayerServicer(t)
		log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
		websocketer := mockService.NewMockWebsocketer(t)

		ctx := t.Context()
		conf, err := config.LoadConfig(ctx)
		require.NoError(t, err)

		rules, err := views.RuleMarkdown("fibbing_it")
		require.NoError(t, err)

		conf.Timings.ShowQuestionScreenFor = time.Millisecond * 1
		conf.Timings.ShowVotingScreenFor = time.Millisecond * 1
		conf.Timings.ShowRevealScreenFor = time.Millisecond * 1
		conf.Timings.ShowScoreScreenFor = time.Millisecond * 1
		conf.Timings.ShowWinnerScreenFor = time.Millisecond * 1

		sub := websockets.NewSubscriber(lobbyService, playerService, roundService, log, websocketer, conf, rules)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		q := websockets.VotingState{
			GameStateID: u,
			Subscriber:  *sub,
		}

		roundService.EXPECT().
			UpdateStateToVoting(mock.Anything, u, mock.AnythingOfType("time.Time")).
			Return(service.VotingState{}, fmt.Errorf("failed to update state"))

		q.Start(ctx)
	})

	t.Run("Should successfully complete reveal state and move to question state", func(t *testing.T) {
		t.Parallel()

		roundService := mockService.NewMockRoundServicer(t)
		lobbyService := mockService.NewMockLobbyServicer(t)
		playerService := mockService.NewMockPlayerServicer(t)
		log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
		websocketer := mockService.NewMockWebsocketer(t)

		ctx := t.Context()
		conf, err := config.LoadConfig(ctx)
		require.NoError(t, err)

		rules, err := views.RuleMarkdown("fibbing_it")
		require.NoError(t, err)

		conf.Timings.ShowQuestionScreenFor = time.Millisecond * 1
		conf.Timings.ShowVotingScreenFor = time.Millisecond * 1
		conf.Timings.ShowRevealScreenFor = time.Millisecond * 1
		conf.Timings.ShowScoreScreenFor = time.Millisecond * 1
		conf.Timings.ShowWinnerScreenFor = time.Millisecond * 1

		sub := websockets.NewSubscriber(lobbyService, playerService, roundService, log, websocketer, conf, rules)

		u, err := uuid.NewV7()
		require.NoError(t, err)
		q := websockets.RevealState{
			GameStateID: u,
			Subscriber:  *sub,
		}

		p1, err := uuid.NewV7()
		require.NoError(t, err)
		p2, err := uuid.NewV7()
		require.NoError(t, err)
		roundService.EXPECT().
			UpdateStateToReveal(mock.Anything, u, mock.AnythingOfType("time.Time")).
			Return(service.RevealRoleState{
				PlayerIDs: []uuid.UUID{p1, p2},
				Round:     1,
				RoundType: "free_form",
				Deadline:  time.Second * 1,
			}, nil)

		playerService.EXPECT().GetPlayerByID(mock.Anything, p1).Return(db.Player{Locale: pgtype.Text{String: "en-US", Valid: true}}, nil)
		playerService.EXPECT().GetPlayerByID(mock.Anything, p2).Return(db.Player{Locale: pgtype.Text{String: "en-US", Valid: true}}, nil)

		websocketer.EXPECT().Publish(mock.Anything, p1, mock.AnythingOfType("[]uint8")).Return(nil)
		websocketer.EXPECT().Publish(mock.Anything, p2, mock.AnythingOfType("[]uint8")).Return(nil)

		roundService.EXPECT().
			UpdateStateToQuestion(mock.Anything, u, mock.AnythingOfType("time.Time"), false).
			Return(service.QuestionState{}, fmt.Errorf("stop state machine here"))

		q.Start(ctx)
		// INFO: Wait for the question state goroutine to spin up
		time.Sleep(10 * time.Millisecond)
	})

	t.Run(
		"Should successfully complete reveal state and move to winner state because final round reached",
		func(t *testing.T) {
			t.Parallel()

			roundService := mockService.NewMockRoundServicer(t)
			lobbyService := mockService.NewMockLobbyServicer(t)
			playerService := mockService.NewMockPlayerServicer(t)
			log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
			websocketer := mockService.NewMockWebsocketer(t)

			ctx := t.Context()
			conf, err := config.LoadConfig(ctx)
			require.NoError(t, err)

			rules, err := views.RuleMarkdown("fibbing_it")
			require.NoError(t, err)

			conf.Timings.ShowQuestionScreenFor = time.Millisecond * 1
			conf.Timings.ShowVotingScreenFor = time.Millisecond * 1
			conf.Timings.ShowRevealScreenFor = time.Millisecond * 1
			conf.Timings.ShowScoreScreenFor = time.Millisecond * 1
			conf.Timings.ShowWinnerScreenFor = time.Millisecond * 1

			sub := websockets.NewSubscriber(lobbyService, playerService, roundService, log, websocketer, conf, rules)

			u, err := uuid.NewV7()
			require.NoError(t, err)
			q := websockets.RevealState{
				GameStateID: u,
				Subscriber:  *sub,
			}

			p1, err := uuid.NewV7()
			require.NoError(t, err)
			p2, err := uuid.NewV7()
			require.NoError(t, err)
			roundService.EXPECT().
				UpdateStateToReveal(mock.Anything, u, mock.AnythingOfType("time.Time")).
				Return(service.RevealRoleState{
					PlayerIDs:    []uuid.UUID{p1, p2},
					Round:        3,
					RoundType:    "most_likely",
					ShouldReveal: true,
					Deadline:     time.Second * 1,
				}, nil)

			playerService.EXPECT().GetPlayerByID(mock.Anything, p1).Return(db.Player{Locale: pgtype.Text{String: "en-US", Valid: true}}, nil)
			playerService.EXPECT().GetPlayerByID(mock.Anything, p2).Return(db.Player{Locale: pgtype.Text{String: "en-US", Valid: true}}, nil)

			websocketer.EXPECT().Publish(mock.Anything, p1, mock.AnythingOfType("[]uint8")).Return(nil)
			websocketer.EXPECT().Publish(mock.Anything, p2, mock.AnythingOfType("[]uint8")).Return(nil)

			roundService.EXPECT().
				UpdateStateToWinner(mock.Anything, u, mock.AnythingOfType("time.Time")).
				Return(service.WinnerState{}, fmt.Errorf("stop state machine here"))

			q.Start(ctx)
			// INFO: Wait for the question state goroutine to spin up
			time.Sleep(10 * time.Millisecond)
		},
	)

	t.Run(
		"Should successfully complete scoring state and move to question state",
		func(t *testing.T) {
			t.Parallel()

			roundService := mockService.NewMockRoundServicer(t)
			lobbyService := mockService.NewMockLobbyServicer(t)
			playerService := mockService.NewMockPlayerServicer(t)
			log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
			websocketer := mockService.NewMockWebsocketer(t)

			ctx := t.Context()
			conf, err := config.LoadConfig(ctx)
			require.NoError(t, err)

			rules, err := views.RuleMarkdown("fibbing_it")
			require.NoError(t, err)

			conf.Timings.ShowQuestionScreenFor = time.Millisecond * 1
			conf.Timings.ShowVotingScreenFor = time.Millisecond * 1
			conf.Timings.ShowRevealScreenFor = time.Millisecond * 1
			conf.Timings.ShowScoreScreenFor = time.Millisecond * 1
			conf.Timings.ShowWinnerScreenFor = time.Millisecond * 1

			scoring := service.Scoring{
				GuessedFibber:      conf.Scoring.GuessFibber,
				FibberEvadeCapture: conf.Scoring.FibberEvadeCapture,
			}

			sub := websockets.NewSubscriber(lobbyService, playerService, roundService, log, websocketer, conf, rules)

			u, err := uuid.NewV7()
			require.NoError(t, err)
			q := websockets.ScoringState{
				GameStateID: u,
				Subscriber:  *sub,
			}

			p1, err := uuid.NewV7()
			require.NoError(t, err)
			p2, err := uuid.NewV7()
			require.NoError(t, err)
			roundService.EXPECT().
				UpdateStateToScore(mock.Anything, u, mock.AnythingOfType("time.Time"), scoring).
				Return(service.ScoreState{
					Players: []service.PlayerWithScoring{
						{
							ID:       p1,
							Nickname: "Player1",
							Avatar:   "avatar1.png",
						},
						{
							ID:       p2,
							Nickname: "Player2",
							Avatar:   "avatar2.png",
						},
					},
					RoundType:   "free_form",
					RoundNumber: 1,
					Deadline:    time.Second * 1,
				}, nil)

			playerService.EXPECT().GetPlayerByID(mock.Anything, p1).Return(db.Player{Locale: pgtype.Text{String: "en-US", Valid: true}}, nil)
			playerService.EXPECT().GetPlayerByID(mock.Anything, p2).Return(db.Player{Locale: pgtype.Text{String: "en-US", Valid: true}}, nil)

			websocketer.EXPECT().Publish(mock.Anything, p1, mock.AnythingOfType("[]uint8")).Return(nil)
			websocketer.EXPECT().Publish(mock.Anything, p2, mock.AnythingOfType("[]uint8")).Return(nil)

			roundService.EXPECT().
				AreAllPlayersAnswerReady(mock.Anything, u).
				Return(false, nil).
				Maybe()

			roundService.EXPECT().
				UpdateStateToQuestion(mock.Anything, u, mock.AnythingOfType("time.Time"), true).
				Return(service.QuestionState{}, fmt.Errorf("stop state machine here"))

			q.Start(ctx)
			// INFO: Wait for the question state goroutine to spin up
			time.Sleep(10 * time.Millisecond)
		},
	)

	t.Run(
		"Should successfully complete winner state and finish the game",
		func(t *testing.T) {
			t.Parallel()

			roundService := mockService.NewMockRoundServicer(t)
			lobbyService := mockService.NewMockLobbyServicer(t)
			playerService := mockService.NewMockPlayerServicer(t)
			log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
			websocketer := mockService.NewMockWebsocketer(t)

			ctx := t.Context()
			conf, err := config.LoadConfig(ctx)
			require.NoError(t, err)

			rules, err := views.RuleMarkdown("fibbing_it")
			require.NoError(t, err)

			conf.Timings.ShowQuestionScreenFor = time.Millisecond * 1
			conf.Timings.ShowVotingScreenFor = time.Millisecond * 1
			conf.Timings.ShowRevealScreenFor = time.Millisecond * 1
			conf.Timings.ShowScoreScreenFor = time.Millisecond * 1
			conf.Timings.ShowWinnerScreenFor = time.Millisecond * 1

			sub := websockets.NewSubscriber(lobbyService, playerService, roundService, log, websocketer, conf, rules)

			u, err := uuid.NewV7()
			require.NoError(t, err)
			q := websockets.WinnerState{
				GameStateID: u,
				Subscriber:  *sub,
			}

			p1, err := uuid.NewV7()
			require.NoError(t, err)
			p2, err := uuid.NewV7()
			require.NoError(t, err)
			roundService.EXPECT().
				UpdateStateToWinner(mock.Anything, u, mock.AnythingOfType("time.Time")).
				Return(service.WinnerState{
					Players: []service.PlayerWithScoring{
						{
							ID:       p1,
							Nickname: "Player1",
							Avatar:   "avatar1.png",
						},
						{
							ID:       p2,
							Nickname: "Player2",
							Avatar:   "avatar2.png",
						},
					},
				}, nil)

			playerService.EXPECT().GetPlayerByID(mock.Anything, p1).Return(db.Player{Locale: pgtype.Text{String: "en-US", Valid: true}}, nil)
			playerService.EXPECT().GetPlayerByID(mock.Anything, p2).Return(db.Player{Locale: pgtype.Text{String: "en-US", Valid: true}}, nil)

			websocketer.EXPECT().Publish(mock.Anything, p1, mock.AnythingOfType("[]uint8")).Return(nil)
			websocketer.EXPECT().Publish(mock.Anything, p2, mock.AnythingOfType("[]uint8")).Return(nil)
			roundService.EXPECT().FinishGame(mock.Anything, u).Return(nil)

			q.Start(ctx)
		},
	)
}
