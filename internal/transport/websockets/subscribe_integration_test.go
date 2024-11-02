package websockets_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/banterbustest"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets"
)

func TestIntegrationSubscribe(t *testing.T) {
	ctx := context.Background()
	db, err := banterbustest.CreateDB(ctx)
	require.NoError(t, err)

	myStore, err := store.NewStore(db)
	require.NoError(t, err)

	userRandomizer := service.NewUserRandomizer()
	lobbyService := service.NewLobbyService(myStore, userRandomizer)
	playerService := service.NewPlayerService(myStore, userRandomizer)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)
	subscriber := websockets.NewSubscriber(lobbyService, playerService, logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := subscriber.Subscribe(r, w)
		require.NoError(t, err)
	}))

	defer server.Close()

	const mainPlayerNickname = "test"
	const otherPlayerNickname = "test1"

	t.Run("Should successfully handle create room message", func(t *testing.T) {
		conn, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn.Close()

		message := map[string]string{
			"game_name":       "fibbing_it",
			"player_nickname": mainPlayerNickname,
			"message_type":    "create_room",
		}

		msg := sendMessage(message, t, conn)

		assert.Contains(t, msg, mainPlayerNickname)
		assert.Contains(t, msg, "Room Code")
	})

	t.Run("Should successfully handle join room message", func(t *testing.T) {
		conn, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn.Close()

		message := map[string]string{
			"game_name":       "fibbing_it",
			"player_nickname": otherPlayerNickname,
			"message_type":    "create_room",
		}

		msg := sendMessage(message, t, conn)

		conn2, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn2.Close()

		roomCode, err := getRoomCode(msg)
		require.NoError(t, err)

		message = map[string]string{
			"room_code":       roomCode,
			"player_nickname": mainPlayerNickname,
			"message_type":    "join_lobby",
		}

		msg = sendMessage(message, t, conn2)
		assert.Contains(t, msg, mainPlayerNickname)
		assert.Contains(t, msg, "Room Code")
	})

	t.Run("Should successfully handle update nickname and generate new avatar", func(t *testing.T) {
		conn, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn.Close()

		message := map[string]string{
			"game_name":       "fibbing_it",
			"player_nickname": otherPlayerNickname,
			"message_type":    "create_room",
		}

		msg := sendMessage(message, t, conn)

		roomCode, err := getRoomCode(msg)
		require.NoError(t, err)

		conn2, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn2.Close()

		// TODO: refactor creating a room
		message = map[string]string{
			"room_code":       roomCode,
			"player_nickname": mainPlayerNickname,
			"message_type":    "join_lobby",
		}
		_ = sendMessage(message, t, conn2)

		message = map[string]string{
			"player_nickname": "test2",
			"message_type":    "update_player_nickname",
		}
		msg = sendMessage(message, t, conn2)
		assert.Contains(t, msg, "test2")

		message = map[string]string{
			"message_type": "generate_new_avatar",
		}
		msg2 := sendMessage(message, t, conn2)
		assert.NotEqual(t, msg, msg2)
	})

	t.Run("Should successfully handle kicking a player", func(t *testing.T) {
		t.Skip("Fix this test")
		conn, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn.Close()

		message := map[string]string{
			"game_name":       "fibbing_it",
			"player_nickname": mainPlayerNickname,
			"message_type":    "create_room",
		}
		msg := sendMessage(message, t, conn)

		roomCode, err := getRoomCode(msg)
		require.NoError(t, err)

		conn2, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn2.Close()

		message = map[string]string{
			"room_code":       roomCode,
			"player_nickname": otherPlayerNickname,
			"message_type":    "join_lobby",
		}
		_ = sendMessage(message, t, conn2)

		// message = map[string]string{
		// 	"player_nickname_to_kick": otherPlayerNickname,
		// 	"room_code":               roomCode,
		// 	"message_type":            "kick_player",
		// }

		// msg = sendMessage(message, t, conn)
		// time.Sleep(300 * time.Millisecond)
		// assert.NotContains(t, msg, otherPlayerNickname, "Player should be kicked and not been in returned HTML.")
	})

	t.Run("Should successfully handle start room message", func(t *testing.T) {
		conn, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn.Close()

		playerNickname := "test1"
		message := map[string]string{
			"game_name":       "fibbing_it",
			"player_nickname": playerNickname,
			"message_type":    "create_room",
		}

		msg := sendMessage(message, t, conn)
		require.NoError(t, err)

		roomCode, err := getRoomCode(msg)
		require.NoError(t, err)

		conn2, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn2.Close()

		playerNickname = "test"
		message = map[string]string{
			"room_code":       roomCode,
			"player_nickname": playerNickname,
			"message_type":    "join_lobby",
		}

		_ = sendMessage(message, t, conn2)

		require.NoError(t, err)

		message = map[string]string{
			"message_type": "toggle_player_is_ready",
		}
		_ = sendMessage(message, t, conn)

		time.Sleep(100 * time.Millisecond)

		message = map[string]string{
			"message_type": "toggle_player_is_ready",
		}
		_ = sendMessage(message, t, conn2)

		time.Sleep(100 * time.Millisecond)

		// TODO: use structs and turn to json
		message = map[string]string{
			"message_type": "start_game",
			"room_code":    roomCode,
		}
		msg = sendMessage(message, t, conn)

		time.Sleep(100 * time.Millisecond)
		assert.Contains(t, msg, playerNickname)
		// TODO: fix this test with actual game start
		// assert.Contains(t, msg, "Round Number 1")
	})
}

func sendMessage(message map[string]string, t *testing.T, conn net.Conn) string {
	jsonMessage, err := json.Marshal(message)
	require.NoError(t, err)

	err = wsutil.WriteClientText(conn, jsonMessage)
	require.NoError(t, err)

	m, _, err := wsutil.ReadServerData(conn)
	require.NoError(t, err)
	msg := string(m)
	return msg
}

func getRoomCode(msg string) (string, error) {
	pattern := `value="([A-Z0-9]{5})"`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	matches := re.FindStringSubmatch(msg)
	if len(matches) < 1 {
		return "", fmt.Errorf("failed to match room code")
	}
	roomCode := matches[1]
	return roomCode, nil
}
