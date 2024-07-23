package main

import (
	"log"

	"gitlab.com/hmajid2301/banterbus/internal/transport"
)

func main() {
	srv := transport.NewHTTPServer()
	err := srv.Start()
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
