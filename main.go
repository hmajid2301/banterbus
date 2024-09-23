package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/pressly/goose/v3"

	// used to connect to sqlite
	_ "modernc.org/sqlite"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/logger"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store"
	transporthttp "gitlab.com/hmajid2301/banterbus/internal/transport/http"
	"gitlab.com/hmajid2301/banterbus/internal/transport/websockets"
)

//go:embed db/migrations/*.sql
var fs embed.FS

func main() {
	var exitCode int

	logger := logger.New()
	ctx := gracefulShutdown(logger)

	err := mainLogic(ctx, logger)
	if err != nil {
		logger.Error("failed to run main logic", slog.Any("error", err))
		exitCode = 1
	}
	defer func() { os.Exit(exitCode) }()
}

func mainLogic(ctx context.Context, logger *slog.Logger) error {
	conf, err := config.LoadConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	db, err := store.GetDB(conf.DBFolder)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	logger.Info("Applying migrations")
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

	subscriber := websockets.NewSubscriber(lobbyService, playerService, logger)
	server := transporthttp.NewServer(subscriber, logger)

	err = server.Serve()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func gracefulShutdown(logger *slog.Logger) context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		logger.Info("system call", slog.Any("oscall", oscall))
		cancel()
	}()

	return ctx
}

func runDBMigrations(db *sql.DB) error {
	goose.SetBaseFS(fs)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Up(db, "db/migrations"); err != nil {
		return err
	}
	return nil
}
