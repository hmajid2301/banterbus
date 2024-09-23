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
	roomServicer := service.NewLobbyService(myStore, userRandomizer)
	playerServicer := service.NewPlayerService(myStore, userRandomizer)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	subscriber := websockets.NewSubscriber(roomServicer, playerServicer, logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := subscriber.Subscribe(context.Background(), r, w)
		require.NoError(t, err)
	}))

	defer server.Close()

	// TODO: this can be just one test?
	t.Run("Should successfully handle create room message", func(t *testing.T) {
		conn, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn.Close()

		playerNickname := "test"
		message := map[string]string{
			"game_name":       "fibbing_it",
			"player_nickname": playerNickname,
			"message_type":    "create_room",
		}

		msg := sendMessage(message, t, conn)

		assert.Contains(t, msg, playerNickname)
		assert.Regexp(t, "Code: [A-Z0-9]{5}", msg)
	})

	t.Run("Should successfully handle join room message", func(t *testing.T) {
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

		pattern := `Code: [A-Z0-9]{5}`

		re, err := regexp.Compile(pattern)
		require.NoError(t, err)

		matches := re.FindAllString(msg, -1)
		var roomCode string
		for _, match := range matches {
			roomCode = match
		}

		conn2, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn2.Close()
		// INFO: remove "Code: " from room code
		roomCode = roomCode[6:]

		playerNickname = "test"
		message = map[string]string{
			"room_code":       roomCode,
			"player_nickname": playerNickname,
			"message_type":    "join_lobby",
		}

		msg = sendMessage(message, t, conn2)
		assert.Contains(t, msg, playerNickname)
		assert.Regexp(t, "Code: [A-Z0-9]{5}", msg)
	})

	t.Run("Should successfully handle update nickname and generate new avatar", func(t *testing.T) {
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

		pattern := `Code: [A-Z0-9]{5}`

		re, err := regexp.Compile(pattern)
		require.NoError(t, err)

		matches := re.FindAllString(msg, -1)
		var roomCode string
		for _, match := range matches {
			roomCode = match
		}

		conn2, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn2.Close()
		// INFO: remove "Code: " from room code
		roomCode = roomCode[6:]

		playerNickname = "test"
		message = map[string]string{
			"room_code":       roomCode,
			"player_nickname": playerNickname,
			"message_type":    "join_lobby",
		}
		msg = sendMessage(message, t, conn2)
		player2ID, err := getPlayerIDFromHTML(msg)
		require.NoError(t, err)

		message = map[string]string{
			"player_id":       player2ID,
			"player_nickname": "test2",
			"message_type":    "update_player_nickname",
		}
		msg = sendMessage(message, t, conn2)
		assert.Contains(t, msg, "test2")

		// TODO: fix this for ci
		// message = map[string]string{
		// 	"player_id":    player2ID,
		// 	"message_type": "generate_new_avatar",
		// }
		// msg2 := sendMessage(message, t, conn2)
		// assert.NotEqual(t, msg, msg2)
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
		player1ID, err := getPlayerIDFromHTML(msg)
		require.NoError(t, err)

		pattern := `Code: [A-Z0-9]{5}`
		re, err := regexp.Compile(pattern)
		require.NoError(t, err)

		matches := re.FindAllString(msg, -1)
		var roomCode string
		for _, match := range matches {
			roomCode = match
		}

		conn2, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn2.Close()
		// INFO: remove "Code: " from room code
		roomCode = roomCode[6:]

		playerNickname = "test"
		message = map[string]string{
			"room_code":       roomCode,
			"player_nickname": playerNickname,
			"message_type":    "join_lobby",
		}

		msg = sendMessage(message, t, conn2)

		player2ID, err := getPlayerIDFromHTML(msg)
		require.NoError(t, err)
		fmt.Println("Player 1 ID:", player2ID, player1ID)

		message = map[string]string{
			"player_id":    player1ID,
			"message_type": "toggle_player_is_ready",
		}
		_ = sendMessage(message, t, conn)

		time.Sleep(100 * time.Millisecond)

		message = map[string]string{
			"player_id":    player2ID,
			"message_type": "toggle_player_is_ready",
		}
		_ = sendMessage(message, t, conn2)

		message = map[string]string{
			"player_id":    player1ID,
			"message_type": "start_game",
		}
		msg = sendMessage(message, t, conn)
		assert.Contains(t, msg, playerNickname)
	})
}

func getPlayerIDFromHTML(msg string) (string, error) {
	re := regexp.MustCompile(`<input class="hidden" name="player_id" value="([^"]+)"`)
	matches := re.FindStringSubmatch(msg)
	var playerID string
	if len(matches) > 1 {
		playerID = matches[1]
	} else {
		return "", fmt.Errorf("Player ID not found")
	}
	return playerID, nil
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
