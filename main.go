package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/invopop/ctxi18n"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdobak/go-xerrors"
	"github.com/pressly/goose/v3"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"

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
		logger := logging.New(slog.LevelInfo, []slog.Attr{})
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
		return xerrors.New("failed to load config", err)
	}
	hostname, err := os.Hostname()
	if err != nil {
		return xerrors.New("failed to fetch hostname", err)
	}

	logger := logging.New(conf.App.LogLevel, []slog.Attr{
		slog.String("app_name", "banterbus"),
		slog.String("node", hostname),
		slog.String("environment", conf.App.Environment),
	})

	// TODO: refactor this
	pgxConfig, err := pgxpool.ParseConfig(conf.DB.URI)
	if err != nil {
		return xerrors.New("failed to parse db uri", err)
	}

	pgxConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxUUID.Register(conn.TypeMap())
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return xerrors.New("failed to setup database", err)
	}
	defer pool.Close()

	otelShutdown, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		return xerrors.New("failed to setup otel", err)
	}

	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	logger.InfoContext(ctx, "applying migrations")
	err = runDBMigrations(pool)
	if err != nil {
		return xerrors.New("failed to run migrations", err)
	}

	str, err := db.NewDB(pool)
	if err != nil {
		return xerrors.New("failed to setup store", err)
	}

	err = ctxi18n.LoadWithDefault(views.Locales, conf.App.DefaultLocale)
	if err != nil {
		return xerrors.New("error loading locales", err)
	}

	userRandomizer := randomizer.NewUserRandomizer()
	lobbyService := service.NewLobbyService(str, userRandomizer, conf.App.DefaultLocale.String())
	playerService := service.NewPlayerService(str, userRandomizer)
	roundService := service.NewRoundService(str, userRandomizer, conf.App.DefaultLocale.String())

	fsys, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return xerrors.New("failed to create embed file system", err)
	}

	redisClient := pubsub.NewRedisClient(conf.Redis.Address)
	subscriber := websockets.NewSubscriber(lobbyService, playerService, roundService, logger, redisClient, conf)

	serverConfig := transporthttp.ServerConfig{
		Host:          conf.Server.Host,
		Port:          conf.Server.Port,
		DefaultLocale: conf.App.DefaultLocale,
	}
	server := transporthttp.NewServer(subscriber, logger, http.FS(fsys), serverConfig)

	timeoutSeconds := 15
	go terminateHandler(ctx, logger, server, timeoutSeconds)
	err = server.Serve(ctx)
	if err != nil {
		return xerrors.New("failed to start server", err)
	}
	return nil
}

// terminateHandler waits for SIGINT or SIGTERM signals and does a graceful shutdown of the HTTP server
// Wait for interrupt signal to gracefully shutdown the server with
// a timeout of 5 seconds.
// kill (no param) default send syscall.SIGTERM
// kill -2 is syscall.SIGINT
// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
func terminateHandler(ctx context.Context, logger *slog.Logger, srv *transporthttp.Server, timeout int) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.InfoContext(ctx, "shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.ErrorContext(ctx, "unexpected error while shutting down server", slog.Any("error", err))
	}
}

func runDBMigrations(pool *pgxpool.Pool) error {
	goose.SetBaseFS(migrations)

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
