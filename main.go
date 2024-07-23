package main

import (
	"log"

	"gitlab.com/hmajid2301/banterbus/internal/transport/ws"
)

func main() {
	srv := ws.NewHTTPServer()
	err := srv.Serve()
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
