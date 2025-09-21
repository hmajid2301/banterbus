package main

import (
	"context"

	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/invopop/ctxi18n"

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

//go:embed static
var staticFiles embed.FS

func main() {
	var exitCode int

	err := mainLogic()
	if err != nil {
		logger := telemetry.NewLogger()
		logger.ErrorContext(context.Background(), "failed to start app", slog.Any("error", err))
		exitCode = 1
	}
	defer func() { os.Exit(exitCode) }()
}

func mainLogic() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := config.LoadConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	telemtryShtudown, err := telemetry.Setup(ctx, conf.App.Environment, conf.App.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to setup otel: %w", err)
	}

	defer func() {
		err = errors.Join(err, telemtryShtudown(ctx))
	}()

	logger := telemetry.NewLogger()

	pool, err := db.NewPool(ctx, conf.DB.URI)
	if err != nil {
		return fmt.Errorf("failed to setup database pool: %w", err)
	}
	defer pool.Close()

	database := db.NewDB(pool, conf.App.Retries, conf.App.BaseDelay)

	err = ctxi18n.LoadWithDefault(views.Locales, conf.App.DefaultLocale)
	if err != nil {
		return fmt.Errorf("error loading locales: %w", err)
	}

	userRandomizer := randomizer.NewUserRandomizer()
	lobbyService := service.NewLobbyService(database, userRandomizer, conf.App.DefaultLocale.String())
	playerService := service.NewPlayerService(database, userRandomizer)
	roundService := service.NewRoundService(database, userRandomizer, conf.App.DefaultLocale.String())
	questionService := service.NewQuestionService(database, userRandomizer, conf.App.DefaultLocale.String())

	fsys, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("failed to create embed file system: %w", err)
	}

	redisClient, err := pubsub.NewRedisClient(conf.Redis.Address, conf.App.Retries)
	if err != nil {
		return fmt.Errorf("failed to create redis client: %w", err)
	}

	rules, err := views.RuleMarkdown(conf.App.DefaultGame)
	if err != nil {
		return fmt.Errorf("failed to convert rules MD to HTML: %w", err)
	}

	subscriber := websockets.NewSubscriber(lobbyService, playerService, roundService, logger, &redisClient, conf, rules)

	var k keyfunc.Keyfunc
	if conf.JWT.JWKSURL != "" {
		storage, err := jwkset.NewStorageFromHTTP(conf.JWT.JWKSURL, jwkset.HTTPClientStorageOptions{Ctx: ctx})
		if err != nil {
			logger.WarnContext(ctx, "failed to setup jwkset storage, JWT validation disabled", slog.Any("error", err))
		} else {
			k, err = keyfunc.New(keyfunc.Options{
				Storage: storage,
			})
			if err != nil {
				logger.WarnContext(ctx, "failed to create keyfunc, JWT validation disabled", slog.Any("error", err))
			}
		}
	}

	serverConfig := transporthttp.ServerConfig{
		Host:          conf.Server.Host,
		Port:          conf.Server.Port,
		DefaultLocale: conf.App.DefaultLocale,
		Environment:   conf.App.Environment,
	}
	var keyFunc func(token *jwt.Token) (interface{}, error)
	if k != nil {
		keyFunc = k.Keyfunc
	}
	server := transporthttp.NewServer(subscriber, logger, http.FS(fsys), keyFunc, questionService, serverConfig)

	go func() {
		logger.InfoContext(
			ctx,
			"starting server",
			slog.String("host", conf.Server.Host),
			slog.Int("port", conf.Server.Port),
		)
		if err := server.Serve(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to serve server", "error", err)
		}
	}()

	timeoutSeconds := 25
	terminateHandler(ctx, logger, server, timeoutSeconds)

	return nil
}

// terminateHandler waits for SIGINT or SIGTERM signals and does a graceful shutdown of the HTTP server
// Wait for interrupt signal to gracefully shutdown the server with
// a timeout of 25 seconds.
// kill (no param) default send syscall.SIGTERM
// kill -2 is syscall.SIGINT
// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
func terminateHandler(ctx context.Context, logger *slog.Logger, srv *transporthttp.Server, timeout int) {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	stop()
	logger.InfoContext(ctx, "shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.ErrorContext(ctx, "unexpected error while shutting down server", slog.Any("error", err))
	}
}
