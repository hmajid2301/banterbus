package main

import (
	"context"
	"database/sql"
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
	"github.com/exaring/otelpgx"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/invopop/ctxi18n"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"

	"gitlab.com/hmajid2301/banterbus/internal/banterbustest"
	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/logging"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/service/randomizer"
	"gitlab.com/hmajid2301/banterbus/internal/store/db"
	"gitlab.com/hmajid2301/banterbus/internal/store/pubsub"
	"gitlab.com/hmajid2301/banterbus/internal/telemetry"
	transporthttp "gitlab.com/hmajid2301/banterbus/internal/transport/http"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets"
	"gitlab.com/hmajid2301/banterbus/internal/views"
)

//go:embed internal/store/db/sqlc/migrations/*.sql
var migrations embed.FS

//go:embed static
var staticFiles embed.FS

func main() {
	var exitCode int

	err := mainLogic()
	if err != nil {
		logger := logging.New()
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

	otelShutdown, err := telemetry.SetupOTelSDK(ctx, conf.App.Environment, conf.App.DisableTelemetry)
	if err != nil {
		return fmt.Errorf("failed to setup otel: %w", err)
	}

	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// TODO: take these values from otel? instrument via otel?
	logger := logging.New()

	// TODO: refactor this
	pgxConfig, err := pgxpool.ParseConfig(conf.DB.URI)
	if err != nil {
		return fmt.Errorf("failed to parse db uri: %w", err)
	}

	pgxConfig.ConnConfig.Tracer = otelpgx.NewTracer()

	pgxConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	defer pool.Close()

	logger.InfoContext(ctx, "applying migrations")
	err = runDBMigrations(pool)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Populate test data if running in test environment
	if conf.App.Environment == "test" {
		logger.InfoContext(ctx, "populating test data")
		err = populateTestData(ctx, pool)
		if err != nil {
			return fmt.Errorf("failed to populate test data: %w", err)
		}
	}

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

	// TODO: should we stop startup if this is failing?
	storage, err := jwkset.NewStorageFromHTTP(conf.JWT.JWKSURL, jwkset.HTTPClientStorageOptions{Ctx: ctx})
	if err != nil {
		return fmt.Errorf("failed to jwkset storage: %w", err)
	}

	k, err := keyfunc.New(keyfunc.Options{
		Storage: storage,
	})
	if err != nil {
		return fmt.Errorf("failed to create keyfunc: %w", err)
	}

	serverConfig := transporthttp.ServerConfig{
		Host:          conf.Server.Host,
		Port:          conf.Server.Port,
		DefaultLocale: conf.App.DefaultLocale,
		Environment:   conf.App.Environment,
	}
	server := transporthttp.NewServer(subscriber, logger, http.FS(fsys), k.Keyfunc, questionService, serverConfig)

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

func runDBMigrations(pool *pgxpool.Pool) error {
	goose.SetBaseFS(migrations)
	goose.WithLogger(goose.NopLogger())

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	cp := pool.Config().ConnConfig.ConnString()
	db, err := sql.Open("pgx/v5", cp)
	if err != nil {
		return err
	}

	err = goose.Up(db, "internal/store/db/sqlc/migrations")
	return err
}

func populateTestData(ctx context.Context, pool *pgxpool.Pool) error {
	return banterbustest.FillWithDummyData(ctx, pool)
}
