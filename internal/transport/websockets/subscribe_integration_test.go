package websockets_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
	db := banterbustest.CreateDB(ctx, t)
	myStore, err := store.NewStore(db)
	require.NoError(t, err)

	userRandomizer := service.NewUserRandomizer()
	roomServicer := service.NewRoomService(myStore, userRandomizer)
	playerServicer := service.NewPlayerService(myStore, userRandomizer)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subscriber := websockets.NewSubscriber(roomServicer, playerServicer, logger)
		err := subscriber.Subscribe(context.Background(), r, w)
		require.NoError(t, err)
	}))

	defer server.Close()
	conn, _, _, err := ws.Dial(context.Background(), fmt.Sprintf("ws://%s", server.Listener.Addr().String()))
	require.NoError(t, err)
	defer conn.Close()

	t.Run("Should successfully handle create room message", func(t *testing.T) {
		playerNickname := "test"
		message := map[string]string{
			"game_name":       "fibbing_it",
			"player_nickname": playerNickname,
			"message_type":    "create_room",
		}

		jsonMessage, err := json.Marshal(message)
		require.NoError(t, err)

		err = wsutil.WriteClientText(conn, jsonMessage)
		require.NoError(t, err)

		m, op, err := wsutil.ReadServerData(conn)
		require.NoError(t, err)
		msg := string(m)

		assert.Equal(t, ws.OpText, op)
		assert.Contains(t, msg, playerNickname)
		assert.Regexp(t, "Code: [A-Z0-9]{5}", msg)
	})

	// t.Run("Should successfully handle join room message", func(t *testing.T) {
	// 	playerNickname := "test1"
	// 	message := map[string]string{
	// 		"game_name":       "fibbing_it",
	// 		"player_nickname": playerNickname,
	// 		"message_type":    "create_room",
	// 	}
	//
	// 	jsonMessage, err := json.Marshal(message)
	// 	require.NoError(t, err)
	//
	// 	err = wsutil.WriteClientText(conn, jsonMessage)
	// 	require.NoError(t, err)
	//
	// 	m, _, err := wsutil.ReadServerData(conn)
	// 	require.NoError(t, err)
	// 	msg := string(m)
	//
	// 	pattern := `Code: [A-Z0-9]{5}`
	//
	// 	re, err := regexp.Compile(pattern)
	// 	if err != nil {
	// 		fmt.Println("Error compiling regex:", err)
	// 		return
	// 	}
	//
	// 	matches := re.FindAllString(msg, -1)
	// 	var roomCode string
	// 	for _, match := range matches {
	// 		roomCode = match
	// 	}
	//
	// 	playerNickname = "test"
	// 	message = map[string]string{
	// 		"room_code":       roomCode,
	// 		"player_nickname": playerNickname,
	// 		"message_type":    "join_room",
	// 	}
	// 	//
	// 	jsonMessage, err = json.Marshal(message)
	// 	require.NoError(t, err)
	//
	// 	err = wsutil.WriteClientText(conn, jsonMessage)
	// 	require.NoError(t, err)
	//
	// 	m, _, err = wsutil.ReadServerData(conn)
	// 	require.NoError(t, err)
	// 	msg = string(m)
	//
	// 	assert.Contains(t, msg, playerNickname)
	// 	assert.Regexp(t, "Code: [A-Z0-9]{5}", msg)
	// })
}
