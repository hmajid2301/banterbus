package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/invopop/ctxi18n"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/banterbustest"
	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/views"
)

func setupSubtest(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := t.Context()
	db, err := banterbustest.CreateDB(ctx)
	require.NoError(t, err)

	return db, func() {
		db.Close()
		banterbustest.RemoveDB(ctx, db)
	}
}

// TODO: use this function
func createRoom(ctx context.Context, srv *service.LobbyService) (service.Lobby, error) {
	newPlayer := service.NewHostPlayer{
		ID: defaultHostPlayerID,
	}
	result, err := srv.Create(ctx, "fibbing_it", newPlayer)
	return result.Lobby, err
}

func lobbyWithTwoPlayers(ctx context.Context, srv *service.LobbyService) (service.Lobby, error) {
	newPlayer := service.NewHostPlayer{
		ID:       defaultHostPlayerID,
		Nickname: defaultHostNickname,
	}
	result, err := srv.Create(ctx, "fibbing_it", newPlayer)
	if err != nil {
		return service.Lobby{}, err
	}

	joinResult, err := srv.Join(ctx, result.Lobby.Code, defaultOtherPlayerID, defaultOtherPlayerNickname)
	return joinResult.Lobby, err
}

func startGame(
	ctx context.Context,
	lobbySrv *service.LobbyService,
	playerSrv *service.PlayerService,
) (service.QuestionState, error) {
	newPlayer := service.NewHostPlayer{
		ID:       defaultHostPlayerID,
		Nickname: defaultHostNickname,
	}
	l, err := lobbySrv.Create(ctx, "fibbing_it", newPlayer)
	if err != nil {
		return service.QuestionState{}, err
	}

	lobby, err := lobbySrv.Join(ctx, l.Lobby.Code, defaultOtherPlayerID, defaultOtherPlayerNickname)
	if err != nil {
		return service.QuestionState{}, err
	}

	_, err = playerSrv.TogglePlayerIsReady(ctx, defaultHostPlayerID)
	if err != nil {
		return service.QuestionState{}, err
	}

	_, err = playerSrv.TogglePlayerIsReady(ctx, defaultOtherPlayerID)
	if err != nil {
		return service.QuestionState{}, err
	}

	questionState, err := lobbySrv.Start(ctx, lobby.Lobby.Code, newPlayer.ID, time.Now().UTC().Add(10*time.Second))
	if err != nil {
		return service.QuestionState{}, err
	}

	if len(questionState.Players) < 2 {
		return service.QuestionState{}, fmt.Errorf(
			"not enough players after game start: got %d, need 2",
			len(questionState.Players),
		)
	}

	return questionState, err
}

func votingState(ctx context.Context,
	lobbyService *service.LobbyService,
	playerService *service.PlayerService,
	roundService *service.RoundService,
) (service.VotingState, error) {
	questionState, err := startGame(ctx, lobbyService, playerService)
	if err != nil {
		return service.VotingState{}, err
	}

	if len(questionState.Players) < 2 {
		return service.VotingState{}, fmt.Errorf(
			"not enough players in question state: got %d, need 2",
			len(questionState.Players),
		)
	}

	err = roundService.SubmitAnswer(ctx, questionState.Players[0].ID, "This is my answer", time.Now())
	if err != nil {
		return service.VotingState{}, err
	}

	err = roundService.SubmitAnswer(
		ctx,
		questionState.Players[1].ID,
		"This is the other players answer",
		time.Now(),
	)
	if err != nil {
		return service.VotingState{}, err
	}

	votingState, err := roundService.UpdateStateToVoting(
		ctx,
		questionState.GameStateID,
		time.Now().Add(120*time.Second),
	)
	return votingState, err
}

func revealState(ctx context.Context,
	lobbyService *service.LobbyService,
	playerService *service.PlayerService,
	roundService *service.RoundService,
) (service.RevealRoleState, error) {
	questionState, err := startGame(ctx, lobbyService, playerService)
	if err != nil {
		return service.RevealRoleState{}, err
	}

	err = roundService.SubmitAnswer(ctx, questionState.Players[0].ID, "This is my answer", time.Now())
	if err != nil {
		return service.RevealRoleState{}, err
	}

	err = roundService.SubmitAnswer(
		ctx,
		questionState.Players[1].ID,
		"This is the other players answer",
		time.Now(),
	)
	if err != nil {
		return service.RevealRoleState{}, err
	}

	votingState, err := roundService.UpdateStateToVoting(
		ctx,
		questionState.GameStateID,
		time.Now().Add(120*time.Second),
	)
	if err != nil {
		return service.RevealRoleState{}, err
	}

	_, err = roundService.SubmitVote(
		ctx,
		votingState.Players[0].ID,
		votingState.Players[1].Nickname,
		time.Now(),
	)
	if err != nil {
		return service.RevealRoleState{}, err
	}
	revealState, err := roundService.UpdateStateToReveal(
		ctx,
		votingState.GameStateID,
		time.Now().Add(120*time.Second),
	)
	return revealState, err
}

func scoreState(ctx context.Context,
	lobbyService *service.LobbyService,
	playerService *service.PlayerService,
	roundService *service.RoundService,
	config config.Config,
) (service.ScoreState, error) {
	questionState, err := startGame(ctx, lobbyService, playerService)
	if err != nil {
		return service.ScoreState{}, err
	}

	err = roundService.SubmitAnswer(ctx, questionState.Players[0].ID, "This is my answer", time.Now())
	if err != nil {
		return service.ScoreState{}, err
	}

	err = roundService.SubmitAnswer(
		ctx,
		questionState.Players[1].ID,
		"This is the other players answer",
		time.Now(),
	)
	if err != nil {
		return service.ScoreState{}, err
	}

	votingState, err := roundService.UpdateStateToVoting(
		ctx,
		questionState.GameStateID,
		time.Now().Add(120*time.Second),
	)
	if err != nil {
		return service.ScoreState{}, err
	}

	_, err = roundService.SubmitVote(
		ctx,
		votingState.Players[0].ID,
		votingState.Players[1].Nickname,
		time.Now(),
	)
	if err != nil {
		return service.ScoreState{}, err
	}
	_, err = roundService.UpdateStateToReveal(
		ctx,
		votingState.GameStateID,
		time.Now().Add(120*time.Second),
	)
	if err != nil {
		return service.ScoreState{}, err
	}

	scoring := service.Scoring{
		GuessedFibber:      config.Scoring.GuessFibber,
		FibberEvadeCapture: config.Scoring.FibberEvadeCapture,
	}
	scoreState, err := roundService.UpdateStateToScore(
		ctx,
		votingState.GameStateID,
		time.Now().Add(120*time.Second),
		scoring)
	return scoreState, err
}

func getI18nCtx() (context.Context, error) {
	ctx := context.Background()
	ctxi18n.LoadWithDefault(views.Locales, "en-GB")
	ctx, err := ctxi18n.WithLocale(ctx, "en-GB")
	return ctx, err
}

// Taken from: https://gist.github.com/StevenACoffman/74347e58e5e0dc4bdf0a79240557c406
func PartialEqual(t require.TestingT, expected, actual any, diffOpts cmp.Option, msgAndArgs ...any) {
	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	if cmp.Equal(expected, actual, diffOpts) {
		return
	}

	diff := cmp.Diff(expected, actual, diffOpts)
	assert.Fail(t, fmt.Sprintf("Not equal: \n"+
		"expected: %s\n"+
		"actual  : %s%s", expected, actual, diff), msgAndArgs...)
}

type tHelper interface {
	Helper()
}
