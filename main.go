package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pressly/goose/v3"

	// used to connect to sqlite
	_ "modernc.org/sqlite"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/logging"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store"
	transporthttp "gitlab.com/hmajid2301/banterbus/internal/transport/http"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

//go:embed static
var staticFiles embed.FS

func main() {
	var exitCode int

	err := mainLogic()
	if err != nil {
		logger := logging.New(slog.LevelInfo, []slog.Attr{})
		logger.Error("failed to start app", slog.Any("error", err))
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
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to fetch hostname: %w", err)
	}

	logger := logging.New(conf.App.LogLevel, []slog.Attr{
		slog.String("app_name", "banterbus"),
		slog.String("node", hostname),
		slog.String("environment", conf.App.Environment),
	})
	db, err := store.GetDB(conf.DBFolder)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	logger.Info("applying migrations")
	err = runDBMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	myStore, err := store.NewStore(db)
	if err != nil {
		return fmt.Errorf("failed to setup store: %w", err)
	}

	userRandomizer := service.NewUserRandomizer()
	lobbyService := service.NewLobbyService(myStore, userRandomizer)
	playerService := service.NewPlayerService(myStore, userRandomizer)

	fsys, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("failed to create embed file system: %w", err)
	}

	subscriber := websockets.NewSubscriber(lobbyService, playerService, logger)
	server := transporthttp.NewServer(subscriber, logger, http.FS(fsys))

	go terminateHandler(logger, server, 60)
	err = server.Serve()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

// terminateHandler waits for SIGINT or SIGTERM signals and does a graceful shutdown of the HTTP server
// Wait for interrupt signal to gracefully shutdown the server with
// a timeout of 5 seconds.
// kill (no param) default send syscall.SIGTERM
// kill -2 is syscall.SIGINT
// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
func terminateHandler(logger *slog.Logger, srv *transporthttp.Server, timeout int) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("unexpected error while shutting down server", slog.Any("error", err))
	}
}

func runDBMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Up(db, "db/migrations"); err != nil {
		return err
	}
	return nil
}
