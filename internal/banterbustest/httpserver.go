package banterbustest

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/mdobak/go-xerrors"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/service/randomizer"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/store/pubsub"
	transporthttp "gitlab.com/hmajid2301/banterbus/internal/transport/http"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets"
	"gitlab.com/hmajid2301/banterbus/internal/views"
)

func NewTestServer() (*httptest.Server, error) {
	ctx := context.Background()
	pool, err := CreateDB(ctx)
	if err != nil {
		return nil, err
	}

	myStore, err := db.NewDB(pool)
	if err != nil {
		return nil, err
	}

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

	redisClient, err := pubsub.NewRedisClient(redisAddr)
	if err != nil {
		return nil, err
	}

	conf, err := config.LoadConfig(ctx)
	if err != nil {
		return nil, err
	}

	showScreenFor := 2
	conf.Timings.ShowScoreScreenFor = time.Second * time.Duration(showScreenFor)
	conf.Timings.ShowQuestionScreenFor = time.Second * time.Duration(showScreenFor)
	conf.Timings.ShowVotingScreenFor = time.Second * time.Duration(showScreenFor)
	conf.Timings.ShowRevealScreenFor = time.Second * time.Duration(showScreenFor)
	conf.App.AutoReconnect = false

	subscriber := websockets.NewSubscriber(lobbyServicer, playerServicer, roundServicer, logger, redisClient, conf)
	err = ctxi18n.LoadWithDefault(views.Locales, "en-GB")
	if err != nil {
		return nil, xerrors.New("error loading locales", err)
	}

	port := 8198
	staticFS := http.Dir("../../static")
	serverConfig := transporthttp.ServerConfig{
		Host:          "localhost",
		Port:          port,
		DefaultLocale: i18n.Code("en-GB"),
		AuthDisabled:  true,
	}

	u, err := url.Parse(conf.JWT.JWKSURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse jwks URL: %w", err)
	}

	storage, err := jwkset.NewStorageFromHTTP(u, jwkset.HTTPClientStorageOptions{Ctx: ctx})
	if err != nil {
		return nil, fmt.Errorf("failed to jwkset storage: %w", err)
	}

	k, err := keyfunc.New(keyfunc.Options{
		Storage: storage,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create keyfunc: %w", err)
	}
	srv := transporthttp.NewServer(subscriber, logger, staticFS, k.Keyfunc, questionServicer, serverConfig)
	server := httptest.NewServer(srv.Server.Handler)

	return server, nil
}

func setupLogger() *slog.Logger {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "DEBUG"
	}

	var level slog.Level
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		log.Fatalf("unknown log level: %s", logLevel)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	if os.Getenv("LOG_DISABLED") == "true" {
		logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
			Level: level,
		}))
	}

	return logger
}
