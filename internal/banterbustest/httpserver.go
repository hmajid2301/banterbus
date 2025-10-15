package banterbustest

import (
	"context"
	"fmt"
	"io"
	"testing"

	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"

	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/service/randomizer"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/store/pubsub"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
	transporthttp "gitlab.com/hmajid2301/banterbus/internal/transport/http"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets"
	"gitlab.com/hmajid2301/banterbus/internal/views"
)

func NewTestServer(t *testing.T) (*httptest.Server, error) {
	pool := NewDB(t)

	baseDelay := time.Millisecond * 100
	retries := 3
	myStore := db.NewDB(pool, retries, baseDelay)

	userRandomizer := randomizer.NewUserRandomizer()
	lobbyServicer := service.NewLobbyService(myStore, userRandomizer, "en-GB")
	playerServicer := service.NewPlayerService(myStore, userRandomizer)
	roundServicer := service.NewRoundService(myStore, userRandomizer, "en-GB")
	questionServicer := service.NewQuestionService(myStore, userRandomizer, "en-GB")
	logger := setupLogger()

	redisAddr := os.Getenv("BANTERBUS_REDIS_ADDRESS")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient, err := pubsub.NewRedisClient(redisAddr, retries)
	if err != nil {
		return nil, err
	}

	conf, err := config.LoadConfig(context.Background())
	if err != nil {
		return nil, err
	}

	showScreenFor := 3
	conf.Timings.ShowScoreScreenFor = time.Second * time.Duration(showScreenFor)
	conf.Timings.ShowQuestionScreenFor = time.Second * time.Duration(showScreenFor)
	conf.Timings.ShowVotingScreenFor = time.Second * time.Duration(showScreenFor)
	conf.Timings.ShowRevealScreenFor = time.Second * time.Duration(showScreenFor)
	conf.App.AutoReconnect = false

	rules, err := views.RuleMarkdown("fibbing_it")
	if err != nil {
		return nil, fmt.Errorf("failed to convert rules MD to HTML: %w", err)
	}

	subscriber := websockets.NewSubscriber(
		lobbyServicer,
		playerServicer,
		roundServicer,
		logger,
		&redisClient,
		conf,
		rules,
		t.Context(),
	)
	err = ctxi18n.LoadWithDefault(views.Locales, "en-GB")
	if err != nil {
		return nil, fmt.Errorf("error loading locales: %w", err)
	}

	port := 8198
	staticFS := http.Dir("../../static")
	serverConfig := transporthttp.ServerConfig{
		Host:          "localhost",
		Port:          port,
		DefaultLocale: i18n.Code("en-GB"),
		AuthDisabled:  true,
	}

	var keyFunc jwt.Keyfunc
	if !serverConfig.AuthDisabled && conf.JWT.JWKSURL != "" {
		storage, err := jwkset.NewStorageFromHTTP(conf.JWT.JWKSURL, jwkset.HTTPClientStorageOptions{Ctx: context.Background()})
		if err != nil {
			return nil, fmt.Errorf("failed to jwkset storage: %w", err)
		}

		k, err := keyfunc.New(keyfunc.Options{
			Storage: storage,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create keyfunc: %w", err)
		}
		keyFunc = k.Keyfunc
	}
	srv := transporthttp.NewServer(subscriber, logger, staticFS, keyFunc, questionServicer, serverConfig)
	server := httptest.NewServer(srv.Server.Handler)

	return server, nil
}

func setupLogger() *slog.Logger {
	if _, ok := os.LookupEnv("BANTERBUS_LOG_DISABLED"); ok {
		handler := slog.NewJSONHandler(io.Discard, nil)
		return slog.New(handler)
	}

	logger := telemetry.NewLogger(slog.LevelInfo)
	return logger
}
