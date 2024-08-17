package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	// used to connect to sqlite
	slogotel "github.com/remychantenay/slog-otel"
	_ "modernc.org/sqlite"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/random"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store"
	transporthttp "gitlab.com/hmajid2301/banterbus/internal/transport/http"
	"gitlab.com/hmajid2301/banterbus/internal/transport/ws"
)

//go:embed db/schema.sql
var ddl string

func main() {
	var exitCode int
	slog.SetDefault(slog.New(slogotel.OtelHandler{
		Next: slog.NewJSONHandler(os.Stdout, nil),
	}))

	logger := slog.Default()
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

	if conf.Environment == "dev" {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	db, err := store.GetDB(conf.DBFolder)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("failed to create database schema: %w", err)
	}

	myStore, err := store.NewStore(db)
	if err != nil {
		return fmt.Errorf("failed to setup store: %w", err)
	}

	userRandomizer := random.NewUserRandomizer()
	roomServicer := service.NewRoomService(myStore, userRandomizer)
	playerServicer := service.NewPlayerService(myStore, userRandomizer)
	subscriber := ws.NewSubscriber(roomServicer, playerServicer, logger)
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
