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
	"github.com/google/uuid"
	"github.com/invopop/ctxi18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/hmajid2301/banterbus/internal/banterbustest"
	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/service/randomizer"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/store/pubsub"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets"
	"gitlab.com/hmajid2301/banterbus/internal/views"
)

func TestIntegrationSubscribe(t *testing.T) {
	ctx := context.Background()
	pool, err := banterbustest.CreateDB(ctx)
	require.NoError(t, err)

	str, err := db.NewDB(pool)
	require.NoError(t, err)

	userRandomizer := randomizer.NewUserRandomizer()
	lobbyService := service.NewLobbyService(str, userRandomizer, "en-GB")
	playerService := service.NewPlayerService(str, userRandomizer)
	roundService := service.NewRoundService(str, userRandomizer, "en-GB")
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	err = ctxi18n.LoadWithDefault(views.Locales, "en-GB")
	if err != nil {
		t.Fatal(err)
	}

	redisAddr := os.Getenv("BANTERBUS_REDIS_ADDRESS")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	conf, err := config.LoadConfig(ctx)
	assert.NoError(t, err)

	redisClient := pubsub.NewRedisClient(redisAddr)
	subscriber := websockets.NewSubscriber(lobbyService, playerService, roundService, logger, redisClient, conf)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		playerID := uuid.Must(uuid.NewV7()).String()
		c := &http.Cookie{
			Name:  "player_id",
			Value: playerID,
		}
		r.AddCookie(c)
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

		msg, err := sendMessage(message, conn)
		require.NoError(t, err)

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

		msg, err := sendMessage(message, conn)
		require.NoError(t, err)

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

		msg, err = sendMessage(message, conn2)
		require.NoError(t, err)

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

		msg, err := sendMessage(message, conn)
		require.NoError(t, err)

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

		_, err = sendMessage(message, conn2)
		require.NoError(t, err)

		message = map[string]string{
			"player_nickname": "test2",
			"message_type":    "update_player_nickname",
		}

		msg, err = sendMessage(message, conn2)
		require.NoError(t, err)
		assert.Contains(t, msg, "test2")

		message = map[string]string{
			"message_type": "generate_new_avatar",
		}
		msg2, err := sendMessage(message, conn2)
		require.NoError(t, err)
		assert.NotEqual(t, msg, msg2)
	})

	t.Run("Should successfully handle kicking a player", func(t *testing.T) {
		conn, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn.Close()

		message := map[string]string{
			"game_name":       "fibbing_it",
			"player_nickname": mainPlayerNickname,
			"message_type":    "create_room",
		}

		msg, err := sendMessage(message, conn)
		require.NoError(t, err)

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
		_, err = sendMessage(message, conn2)
		require.NoError(t, err)

		message = map[string]string{
			"player_nickname_to_kick": otherPlayerNickname,
			"room_code":               roomCode,
			"message_type":            "kick_player",
		}

		_, err = sendMessage(message, conn)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		m, _, err := wsutil.ReadServerData(conn2)
		require.NoError(t, err)
		msg = string(m)
		assert.NotContains(t, msg, otherPlayerNickname, "Player should be kicked and not been in returned HTML.")
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

		msg, err := sendMessage(message, conn)
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

		_, err = sendMessage(message, conn2)
		require.NoError(t, err)

		message = map[string]string{
			"message_type": "toggle_player_is_ready",
		}
		_, err = sendMessage(message, conn)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		message = map[string]string{
			"message_type": "toggle_player_is_ready",
		}
		_, err = sendMessage(message, conn2)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// TODO: use structs and turn to json
		message = map[string]string{
			"message_type": "start_game",
			"room_code":    roomCode,
		}
		_, err = sendMessage(message, conn)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		_, err = readMessage(conn)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		msg2, err := readMessage(conn)
		require.NoError(t, err)

		assert.Contains(t, msg2, "Round 1 / 3")
	})

	t.Run("Should successfully handle submit answer message", func(t *testing.T) {
		conn, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
		require.NoError(t, err)
		defer conn.Close()

		playerNickname := "test1"
		message := map[string]string{
			"game_name":       "fibbing_it",
			"player_nickname": playerNickname,
			"message_type":    "create_room",
		}

		msg, err := sendMessage(message, conn)
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

		_, err = sendMessage(message, conn2)
		require.NoError(t, err)

		message = map[string]string{
			"message_type": "toggle_player_is_ready",
		}
		_, err = sendMessage(message, conn)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		message = map[string]string{
			"message_type": "toggle_player_is_ready",
		}
		_, err = sendMessage(message, conn2)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		message = map[string]string{
			"message_type": "start_game",
			"room_code":    roomCode,
		}
		_, err = sendMessage(message, conn)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		_, err = readMessage(conn)
		require.NoError(t, err)

		message = map[string]string{
			"message_type": "submit_answer",
			"answer":       "this is my answer",
		}
		_, err = sendMessage(message, conn)
		require.NoError(t, err)
	})
}

// TODO: see if can test reconnect logic here, we have an issue with not being able to set a cookie for a connection.
// So we get a new player id with each connection.
// func TestIntegrationReconnect(t *testing.T) {
// 	ctx := context.Background()
// 	db, err := banterbustest.CreateDB(ctx)
// 	require.NoError(t, err)
//
// 	myStore, err := store.NewStore(db)
// 	require.NoError(t, err)
//
// 	userRandomizer := service.NewUserRandomizer()
// 	lobbyService := service.NewLobbyService(myStore, userRandomizer)
// 	playerService := service.NewPlayerService(myStore, userRandomizer)
// 	roundService := service.NewRoundService(myStore)
// 	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
// 	logger := slog.New(handler)
// 	subscriber := websockets.NewSubscriber(lobbyService, playerService, roundService, logger)
//
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		playerID := uuid.Must(uuid.NewV7()).String()
// 		c := &http.Cookie{
// 			Name:  "player_id",
// 			Value: playerID,
// 		}
// 		r.AddCookie(c)
// 		err := subscriber.Subscribe(r, w)
// 		require.NoError(t, err)
// 	}))
//
// 	defer server.Close()
//
// 	t.Run("Should successfully reconnect player to lobby with a single player", func(t *testing.T) {
// 		conn, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
// 		require.NoError(t, err)
// 		defer conn.Close()
//
// 		message := map[string]string{
// 			"game_name":       "fibbing_it",
// 			"player_nickname": "test1",
// 			"message_type":    "create_room",
// 		}
//
// 		msg, err := sendMessage(message, conn)
// 		require.NoError(t, err)
//
// 		roomCode, err := getRoomCode(msg)
// 		require.NoError(t, err)
//
// 		// INFO: Force reconnect, but uses cookie to enable us to reconnect using our player ID.
// 		conn.Close()
// 		conn, _, _, err = ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
// 		require.NoError(t, err)
// 		defer conn.Close()
//
// 		msg, err = readMessage(conn)
// 		require.NoError(t, err)
// 		assert.Contains(t, msg, "test1")
// 		assert.Contains(t, msg, roomCode)
// 	})
//
// 	t.Run("Should successfully reconnect player to lobby with a multiple player", func(t *testing.T) {
// 		conn, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
// 		require.NoError(t, err)
// 		defer conn.Close()
//
// 		message := map[string]string{
// 			"game_name":       "fibbing_it",
// 			"player_nickname": "test1",
// 			"message_type":    "create_room",
// 		}
//
// 		msg, err := sendMessage(message, conn)
// 		require.NoError(t, err)
//
// 		roomCode, err := getRoomCode(msg)
// 		require.NoError(t, err)
//
// 		conn2, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
// 		require.NoError(t, err)
// 		defer conn2.Close()
//
// 		message = map[string]string{
// 			"room_code":       roomCode,
// 			"player_nickname": "test2",
// 			"message_type":    "join_lobby",
// 		}
//
// 		_, err = sendMessage(message, conn2)
// 		require.NoError(t, err)
//
// 		time.Sleep(100 * time.Millisecond)
//
// 		// INFO: Force reconnect, but uses cookie to enable us to reconnect using our player ID.
// 		conn.Close()
// 		conn, _, _, err = ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
// 		require.NoError(t, err)
// 		defer conn.Close()
//
// 		msg, err = readMessage(conn)
// 		require.NoError(t, err)
// 		assert.Contains(t, msg, "test1")
// 		assert.Contains(t, msg, "test2")
// 		assert.Contains(t, msg, roomCode)
// 	})
// }

func sendMessage(message map[string]string, conn net.Conn) (string, error) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return "", err
	}

	err = wsutil.WriteClientText(conn, jsonMessage)
	if err != nil {
		return "", err
	}

	return readMessage(conn)
}

func readMessage(conn net.Conn) (string, error) {
	m, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		return "", err
	}
	msg := string(m)
	return msg, nil
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
