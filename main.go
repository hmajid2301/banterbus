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
	"gitlab.com/hmajid2301/banterbus/internal/recovery"
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
		logger := telemetry.NewLogger(slog.LevelInfo)
		logger.ErrorContext(context.Background(), "failed to start app", slog.Any("error", err))
		exitCode = 1
	}
	defer func() { os.Exit(exitCode) }()
}

func mainLogic() error {
	// INFO: separate shutdown context allows canceling state machines before HTTP shutdown
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

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

	// Initialize metrics
	if err := telemetry.InitializeMetrics(ctx); err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	logLevel := telemetry.NewLogLevelFromMinSev(conf.App.LogLevel)
	logger := telemetry.NewLogger(logLevel)

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

	subscriber := websockets.NewSubscriber(lobbyService, playerService, roundService, logger, &redisClient, conf, rules, shutdownCtx)

	recoveryManager := recovery.NewManager(database, subscriber, subscriber, roundService, logger)
	go func() {
		if err := recoveryManager.RecoverActiveGames(ctx); err != nil {
			logger.WarnContext(ctx, "failed to recover active games", slog.Any("error", err))
		}
	}()

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

	terminateHandler(shutdownCtx, shutdownCancel, logger, server, subscriber)

	return nil
}

func terminateHandler(
	shutdownCtx context.Context,
	shutdownCancel context.CancelFunc,
	logger *slog.Logger,
	srv *transporthttp.Server,
	subscriber *websockets.Subscriber,
) {
	ctx, stop := signal.NotifyContext(shutdownCtx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	stop()
	shutdownStart := time.Now()
	logger.InfoContext(ctx, "received shutdown signal, starting graceful shutdown")

	shutdownCancel()
	logger.InfoContext(ctx, "signaled all state machines to stop")

	stateMachineTimeout := 5 * time.Second
	stateMachineStart := time.Now()
	allCompleted := subscriber.WaitForStateMachines(ctx, stateMachineTimeout)
	stateMachineDuration := time.Since(stateMachineStart)

	if !allCompleted {
		logger.WarnContext(ctx, "forcing cancellation of remaining state machines")
		subscriber.CancelAllStateMachines(ctx)
	}

	totalTimeout := 25 * time.Second
	elapsed := time.Since(shutdownStart)
	remainingTimeout := totalTimeout - elapsed
	if remainingTimeout < 1*time.Second {
		remainingTimeout = 1 * time.Second
	}

	httpShutdownCtx, cancel := context.WithTimeout(context.Background(), remainingTimeout)
	defer cancel()

	logger.InfoContext(ctx, "shutting down HTTP server",
		slog.Duration("timeout", remainingTimeout))

	if err := srv.Shutdown(httpShutdownCtx); err != nil {
		logger.ErrorContext(ctx, "unexpected error while shutting down server", slog.Any("error", err))
	} else {
		logger.InfoContext(ctx, "server shutdown completed successfully")
	}

	totalDuration := time.Since(shutdownStart)
	logger.InfoContext(ctx, "shutdown completed",
		slog.Bool("graceful", allCompleted),
		slog.Duration("state_machine_duration", stateMachineDuration),
		slog.Duration("total_duration", totalDuration))
}
