package websockets_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets"
)

func TestCreateRoomValidation(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully validate valid create room", func(t *testing.T) {
		t.Parallel()
		room := websockets.CreateRoom{
			GameName:       "Test Game",
			PlayerNickname: "TestPlayer",
		}

		err := room.Validate()
		assert.NoError(t, err)
	})

	t.Run("Should reject empty game name", func(t *testing.T) {
		t.Parallel()
		room := websockets.CreateRoom{
			GameName:       "",
			PlayerNickname: "TestPlayer",
		}

		err := room.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "game_name is required")
	})

	t.Run("Should reject empty player nickname", func(t *testing.T) {
		t.Parallel()
		room := websockets.CreateRoom{
			GameName:       "Test Game",
			PlayerNickname: "",
		}

		err := room.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player_nickname is required")
	})

	t.Run("Should reject game name that is too long", func(t *testing.T) {
		t.Parallel()
		room := websockets.CreateRoom{
			GameName:       "This is a very long game name that exceeds the limit",
			PlayerNickname: "TestPlayer",
		}

		err := room.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be <= 50 characters")
	})

	t.Run("Should reject nickname that is too long", func(t *testing.T) {
		t.Parallel()
		room := websockets.CreateRoom{
			GameName:       "Test Game",
			PlayerNickname: "ThisNicknameIsWayTooLongForTheSystem",
		}

		err := room.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be <= 30 characters")
	})
}

func TestJoinLobbyValidation(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully validate valid join lobby", func(t *testing.T) {
		t.Parallel()
		join := websockets.JoinLobby{
			PlayerNickname: "TestPlayer",
			RoomCode:       "ABC12",
		}

		err := join.Validate()
		assert.NoError(t, err)
	})

	t.Run("Should reject empty room code", func(t *testing.T) {
		t.Parallel()
		join := websockets.JoinLobby{
			PlayerNickname: "TestPlayer",
			RoomCode:       "",
		}

		err := join.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "room_code is required")
	})

	t.Run("Should reject empty player nickname", func(t *testing.T) {
		t.Parallel()
		join := websockets.JoinLobby{
			PlayerNickname: "",
			RoomCode:       "ABC12",
		}

		err := join.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player_nickname is required")
	})

	t.Run("Should reject room code that is too long", func(t *testing.T) {
		t.Parallel()
		join := websockets.JoinLobby{
			PlayerNickname: "TestPlayer",
			RoomCode:       "VERYLONGROOMCODE",
		}

		err := join.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be <= 10 characters")
	})
}

func TestStartGameValidation(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully validate valid start game", func(t *testing.T) {
		t.Parallel()
		start := websockets.StartGame{RoomCode: "ABC12"}

		err := start.Validate()
		assert.NoError(t, err)
	})

	t.Run("Should reject empty room code", func(t *testing.T) {
		t.Parallel()
		start := websockets.StartGame{RoomCode: ""}

		err := start.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "room_code is required")
	})
}

func TestUpdateNicknameValidation(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully validate valid update nickname", func(t *testing.T) {
		t.Parallel()
		update := websockets.UpdateNickname{PlayerNickname: "NewNickname"}

		err := update.Validate()
		assert.NoError(t, err)
	})

	t.Run("Should reject empty nickname", func(t *testing.T) {
		t.Parallel()
		update := websockets.UpdateNickname{PlayerNickname: ""}

		err := update.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player_nickname is required")
	})

	t.Run("Should reject nickname that is too long", func(t *testing.T) {
		t.Parallel()
		update := websockets.UpdateNickname{PlayerNickname: "ThisNicknameIsWayTooLongForTheSystem"}

		err := update.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be <= 30 characters")
	})
}

func TestKickPlayerValidation(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully validate valid kick player", func(t *testing.T) {
		t.Parallel()
		kick := websockets.KickPlayer{
			RoomCode:             "ABC12",
			PlayerNicknameToKick: "PlayerToKick",
		}

		err := kick.Validate()
		assert.NoError(t, err)
	})

	t.Run("Should reject empty room code", func(t *testing.T) {
		t.Parallel()
		kick := websockets.KickPlayer{
			RoomCode:             "",
			PlayerNicknameToKick: "PlayerToKick",
		}

		err := kick.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "room_code is required")
	})

	t.Run("Should reject empty player nickname to kick", func(t *testing.T) {
		t.Parallel()
		kick := websockets.KickPlayer{
			RoomCode:             "ABC12",
			PlayerNicknameToKick: "",
		}

		err := kick.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player_nickname_to_kick is required")
	})
}

func TestSubmitAnswerValidation(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully validate valid submit answer", func(t *testing.T) {
		t.Parallel()
		submit := websockets.SubmitAnswer{Answer: "My answer"}

		err := submit.Validate()
		assert.NoError(t, err)
	})

	t.Run("Should reject empty answer", func(t *testing.T) {
		t.Parallel()
		submit := websockets.SubmitAnswer{Answer: ""}

		err := submit.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "answer is required")
	})
}

func TestSubmitVoteValidation(t *testing.T) {
	t.Parallel()

	t.Run("Should successfully validate valid submit vote", func(t *testing.T) {
		t.Parallel()
		vote := websockets.SubmitVote{VotedPlayerNickname: "PlayerToVoteFor"}

		err := vote.Validate()
		assert.NoError(t, err)
	})

	t.Run("Should reject empty voted player nickname", func(t *testing.T) {
		t.Parallel()
		vote := websockets.SubmitVote{VotedPlayerNickname: ""}

		err := vote.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "player nickname is required")
	})
}
