package service

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"gitlab.com/hmajid2301/banterbus/internal/store/db"
)

type Storer interface {
	db.Querier
	CreateRoom(ctx context.Context, arg db.CreateRoomArgs) error
	AddPlayerToRoom(ctx context.Context, arg db.AddPlayerToRoomArgs) error
	StartGame(ctx context.Context, arg db.StartGameArgs) error
	NewRound(ctx context.Context, arg db.NewRoundArgs) error
	AddScores(ctx context.Context, arg db.NewScoresArgs) error
	CreateQuestionWithTranslation(ctx context.Context, arg db.CreateQuestionArgs) (uuid.UUID, error)
	UpdateNicknameWithPlayers(ctx context.Context, arg db.UpdateNicknameArgs) (db.UpdateNicknameResult, error)
	GenerateNewAvatarWithPlayers(ctx context.Context, arg db.GenerateNewAvatarArgs) (db.GenerateNewAvatarResult, error)
	TogglePlayerReadyWithPlayers(ctx context.Context, arg db.TogglePlayerIsReadyArgs) (db.TogglePlayerIsReadyResult, error)
	JoinRoom(ctx context.Context, arg db.JoinRoomArgs) (db.JoinRoomResult, error)
	UpdateStateToVoting(ctx context.Context, arg db.UpdateStateToVotingArgs) (db.UpdateStateToVotingResult, error)
	UpdateStateToReveal(ctx context.Context, arg db.UpdateStateToRevealArgs) (db.UpdateStateToRevealResult, error)
	UpdateStateToScore(ctx context.Context, arg db.UpdateStateToScoreArgs) (db.UpdateStateToScoreResult, error)
	UpdateStateToQuestion(ctx context.Context, arg db.UpdateStateToQuestionArgs) (db.UpdateStateToQuestionResult, error)
	ExecuteTransactionWithRetry(ctx context.Context, fn func(*db.Queries) error) error
}

type Randomizer interface {
	GetNickname() string
	GetAvatar(nickname string) string
	GetRoomCode() string
	GetID() (uuid.UUID, error)
	GetFibberIndex(playersLen int) int
}

func getLobbyPlayers(playerRows []db.GetAllPlayersInRoomRow, roomCode string) Lobby {
	var players []LobbyPlayer
	for _, player := range playerRows {
		isHost := false
		if player.ID == player.HostPlayer {
			isHost = true
		}

		p := LobbyPlayer{
			ID:       player.ID,
			Nickname: player.Nickname,
			Avatar:   player.Avatar,
			IsReady:  player.IsReady.Bool,
			IsHost:   isHost,
		}

		players = append(players, p)
	}

	room := Lobby{
		Code:    roomCode,
		Players: players,
	}
	return room
}

const FibberRole = "fibber"
const NormalRole = "normal"
