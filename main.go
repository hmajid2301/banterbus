package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/signal"

	// used to connect to sqlite
	_ "modernc.org/sqlite"

	"gitlab.com/hmajid2301/banterbus/internal/config"
	"gitlab.com/hmajid2301/banterbus/internal/random"
	"gitlab.com/hmajid2301/banterbus/internal/service"
	"gitlab.com/hmajid2301/banterbus/internal/store"
	"gitlab.com/hmajid2301/banterbus/internal/transport/ws"
)

//go:embed db/schema.sql
var ddl string

func main() {
	var exitCode int
	ctx := gracefulShutdown()
	defer func() { os.Exit(exitCode) }()

	db, err := getDB(ctx, ddl)
	if err != nil {
		log.Fatal("failed to setup database: ", err)
		exitCode = 1
		return
	}

	myStore, err := store.NewStore(db)
	if err != nil {
		log.Fatal("failed to setup store: ", err)
		exitCode = 1
		return
	}

	userRandomizer := random.NewUserRandomizer()
	roomService := service.NewRoomService(myStore, userRandomizer)
	roomRandomizer := random.NewRoomRandomizer()
	srv := ws.NewHTTPServer(roomService, roomRandomizer)

	err = srv.Serve()
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
func getDB(ctx context.Context, ddl string) (*sql.DB, error) {
	conf, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := store.GetDB(conf.DBFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return nil, fmt.Errorf("failed to create database schema: %w", err)
	}
	return db, nil
}

func gracefulShutdown() context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		cancel()
	}()

	return ctx
}
