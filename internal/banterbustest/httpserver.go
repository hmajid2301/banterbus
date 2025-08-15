package banterbustest

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/mdobak/go-xerrors"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/logging"
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

	conf, err := config.LoadConfig(ctx)
	if err != nil {
		return nil, err
	}

	showScreenFor := 3
	conf.Timings.ShowScoreScreenFor = time.Second * time.Duration(showScreenFor)
	conf.Timings.ShowQuestionScreenFor = time.Second * time.Duration(showScreenFor)
	conf.Timings.ShowVotingScreenFor = time.Second * time.Duration(showScreenFor)
	conf.Timings.ShowRevealScreenFor = time.Second * time.Duration(showScreenFor)
	conf.App.AutoReconnect = false

	rules, err := views.RuleMarkdown()
	if err != nil {
		return nil, fmt.Errorf("failed to convert rules MD to HTML: %w", err)
	}

	subscriber := websockets.NewSubscriber(
		lobbyServicer,
		playerServicer,
		roundServicer,
		logger,
		redisClient,
		conf,
		rules,
	)
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

	var keyFunc jwt.Keyfunc
	if !serverConfig.AuthDisabled && conf.JWT.JWKSURL != "" {
		storage, err := jwkset.NewStorageFromHTTP(conf.JWT.JWKSURL, jwkset.HTTPClientStorageOptions{Ctx: ctx})
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
	logLevel := os.Getenv("BANTERBUS_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
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

	if _, ok := os.LookupEnv("BANTERBUS_LOG_DISABLED"); ok {
		handler := slog.NewJSONHandler(io.Discard, nil)
		return slog.New(handler)
	}

	logger := logging.New(level, []slog.Attr{})
	return logger
}
