package statemachine_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/invopop/ctxi18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/banterbustest"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/service/randomizer"
	"gitlab.com/hmajid2301/banterbus/internal/statemachine"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/views"
)

type noopClientUpdater struct{}

func (n *noopClientUpdater) UpdateClientsAboutQuestion(ctx context.Context, questionState service.QuestionState, showModal bool) error {
	return nil
}

func (n *noopClientUpdater) UpdateClientsAboutVoting(ctx context.Context, votingState service.VotingState) error {
	return nil
}

func (n *noopClientUpdater) UpdateClientsAboutReveal(ctx context.Context, revealState service.RevealRoleState) error {
	return nil
}

func (n *noopClientUpdater) UpdateClientsAboutScore(ctx context.Context, scoringState service.ScoreState) error {
	return nil
}

func (n *noopClientUpdater) UpdateClientsAboutWinner(ctx context.Context, winnerState service.WinnerState) error {
	return nil
}

type noopTransitioner struct{}

func (n *noopTransitioner) StartStateMachine(ctx context.Context, gameStateID uuid.UUID, state statemachine.State) {
}

type roundServiceAdapter struct {
	*service.RoundService
}

type testServices struct {
	roundService  *service.RoundService
	lobbyService  *service.LobbyService
	playerService *service.PlayerService
}

func setupIntegrationTest(t *testing.T) (*testServices, *statemachine.StateDependencies, func()) {
	t.Helper()

	pool := banterbustest.NewDB(t)
	cleanup := func() {
		pool.Close()
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	baseDelay := time.Millisecond * 100
	storer := db.NewDB(pool, 3, baseDelay)
	rand := randomizer.NewUserRandomizer()

	roundService := service.NewRoundService(storer, rand, "en-GB")
	lobbyService := service.NewLobbyService(storer, rand, "en-GB")
	playerService := service.NewPlayerService(storer, rand)

	services := &testServices{
		roundService:  roundService,
		lobbyService:  lobbyService,
		playerService: playerService,
	}

	timings := statemachine.Timings{
		ShowQuestionScreenFor: 50 * time.Millisecond,
		ShowVotingScreenFor:   50 * time.Millisecond,
		ShowRevealScreenFor:   50 * time.Millisecond,
		ShowScoreScreenFor:    50 * time.Millisecond,
		ShowWinnerScreenFor:   50 * time.Millisecond,
	}

	scoring := service.Scoring{
		GuessedFibber:      100,
		FibberEvadeCapture: 50,
	}

	deps := &statemachine.StateDependencies{
		RoundService:  &roundServiceAdapter{roundService},
		ClientUpdater: &noopClientUpdater{},
		Transitioner:  &noopTransitioner{},
		Logger:        logger,
		Timings:       timings,
		Scoring:       scoring,
	}

	return services, deps, cleanup
}

func createTestLobby(t *testing.T, ctx context.Context, services *testServices) (string, uuid.UUID, uuid.UUID) {
	t.Helper()

	hostID := uuid.Must(uuid.NewV4())
	otherID := uuid.Must(uuid.NewV4())

	ctxi18n.LoadWithDefault(views.Locales, "en-GB")
	ctx, err := ctxi18n.WithLocale(ctx, "en-GB")
	require.NoError(t, err)

	newPlayer := service.NewHostPlayer{
		ID:       hostID,
		Nickname: "Host Player",
	}
	lobby, err := services.lobbyService.Create(ctx, "fibbing_it", newPlayer)
	require.NoError(t, err)

	_, err = services.lobbyService.Join(ctx, lobby.Lobby.Code, otherID, "Other Player")
	require.NoError(t, err)

	_, err = services.playerService.TogglePlayerIsReady(ctx, hostID)
	require.NoError(t, err)

	_, err = services.playerService.TogglePlayerIsReady(ctx, otherID)
	require.NoError(t, err)

	return lobby.Lobby.Code, hostID, otherID
}

func createTestGame(t *testing.T, ctx context.Context, services *testServices) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()

	code, hostID, otherID := createTestLobby(t, ctx, services)

	questionState, err := services.lobbyService.Start(ctx, code, hostID, time.Now().UTC().Add(10*time.Second))
	require.NoError(t, err)

	return questionState.GameStateID, hostID, otherID
}

func TestIntegrationStateMachine(t *testing.T) {
	t.Parallel()

	t.Run("question transitions to voting after timeout", func(t *testing.T) {
		t.Parallel()

		services, deps, cleanup := setupIntegrationTest(t)
		defer cleanup()

		ctx := context.Background()
		gameStateID, _, _ := createTestGame(t, ctx, services)

		questionState, err := statemachine.NewQuestionState(gameStateID, false, deps)
		require.NoError(t, err)

		err = questionState.Start(ctx)
		require.NoError(t, err)
	})

	t.Run("voting transitions to reveal after timeout", func(t *testing.T) {
		t.Parallel()

		services, deps, cleanup := setupIntegrationTest(t)
		defer cleanup()

		ctx := context.Background()
		gameStateID, hostID, otherID := createTestGame(t, ctx, services)

		err := services.roundService.SubmitAnswer(ctx, hostID, "Answer 1", time.Now())
		require.NoError(t, err)
		err = services.roundService.SubmitAnswer(ctx, otherID, "Answer 2", time.Now())
		require.NoError(t, err)

		votingState, err := statemachine.NewVotingState(gameStateID, deps)
		require.NoError(t, err)

		err = votingState.Start(ctx)
		require.NoError(t, err)
	})

	t.Run("reveal transitions to question when not final round", func(t *testing.T) {
		t.Parallel()

		services, deps, cleanup := setupIntegrationTest(t)
		defer cleanup()

		ctx := context.Background()
		gameStateID, hostID, otherID := createTestGame(t, ctx, services)

		err := services.roundService.SubmitAnswer(ctx, hostID, "Answer 1", time.Now())
		require.NoError(t, err)
		err = services.roundService.SubmitAnswer(ctx, otherID, "Answer 2", time.Now())
		require.NoError(t, err)

		_, err = services.roundService.UpdateStateToVoting(ctx, gameStateID, time.Now().Add(10*time.Second))
		require.NoError(t, err)

		revealState, err := statemachine.NewRevealState(gameStateID, deps)
		require.NoError(t, err)

		err = revealState.Start(ctx)
		assert.NoError(t, err)
	})

	t.Run("scoring transitions to question", func(t *testing.T) {
		t.Parallel()

		services, deps, cleanup := setupIntegrationTest(t)
		defer cleanup()

		ctx := context.Background()
		gameStateID, hostID, otherID := createTestGame(t, ctx, services)

		err := services.roundService.SubmitAnswer(ctx, hostID, "Answer 1", time.Now())
		require.NoError(t, err)
		err = services.roundService.SubmitAnswer(ctx, otherID, "Answer 2", time.Now())
		require.NoError(t, err)

		votingState, err := services.roundService.UpdateStateToVoting(ctx, gameStateID, time.Now().Add(10*time.Second))
		require.NoError(t, err)

		_, err = services.roundService.SubmitVote(ctx, votingState.Players[0].ID, votingState.Players[1].Nickname, time.Now())
		require.NoError(t, err)

		_, err = services.roundService.UpdateStateToReveal(ctx, gameStateID, time.Now().Add(10*time.Second))
		require.NoError(t, err)

		scoringState, err := statemachine.NewScoringState(gameStateID, deps)
		require.NoError(t, err)

		err = scoringState.Start(ctx)
		assert.NoError(t, err)
	})

	t.Run("context cancellation stops state machine", func(t *testing.T) {
		t.Parallel()

		services, deps, cleanup := setupIntegrationTest(t)
		defer cleanup()

		ctx, cancel := context.WithCancel(context.Background())
		gameStateID, _, _ := createTestGame(t, context.Background(), services)

		questionState, err := statemachine.NewQuestionState(gameStateID, false, deps)
		require.NoError(t, err)

		cancel()

		err = questionState.Start(ctx)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("winner state finishes game", func(t *testing.T) {
		t.Parallel()

		services, deps, cleanup := setupIntegrationTest(t)
		defer cleanup()

		ctx := context.Background()
		gameStateID, hostID, otherID := createTestGame(t, ctx, services)

		err := services.roundService.SubmitAnswer(ctx, hostID, "Answer 1", time.Now())
		require.NoError(t, err)
		err = services.roundService.SubmitAnswer(ctx, otherID, "Answer 2", time.Now())
		require.NoError(t, err)

		votingState, err := services.roundService.UpdateStateToVoting(ctx, gameStateID, time.Now().Add(10*time.Second))
		require.NoError(t, err)

		_, err = services.roundService.SubmitVote(ctx, votingState.Players[0].ID, votingState.Players[1].Nickname, time.Now())
		require.NoError(t, err)

		_, err = services.roundService.UpdateStateToReveal(ctx, gameStateID, time.Now().Add(10*time.Second))
		require.NoError(t, err)

		winnerState, err := statemachine.NewWinnerState(gameStateID, deps)
		require.NoError(t, err)

		err = winnerState.Start(ctx)
		assert.NoError(t, err)
	})
}
