package websockets

import (
	"fmt"
	"log"
	"sync"
)

type room struct {
	clients     map[string]*client
	register    chan *client
	unregister  chan *client
	broadcast   chan []byte
	clientMutex sync.Mutex
	quit        chan bool
}

func NewRoom() *room {
	return &room{
		clients:    make(map[string]*client),
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan []byte),
		quit:       make(chan bool),
	}
}

func (r *room) runRoom() {
	log.Println("starting room")
	for {
		select {

		case client := <-r.register:
			r.addClient(client)

		case client := <-r.unregister:
			r.removeClient(client.playerID)

		case message := <-r.broadcast:
			r.broadcastMessage(message)

		case <-r.quit:
			log.Println("stopping room")
			return
		}
	}
}

func (r *room) addClient(client *client) {
	log.Println("client added", client)
	r.clientMutex.Lock()
	r.clients[client.playerID] = client
	r.clientMutex.Unlock()
}

func (r *room) removeClient(playerID string) {
	log.Println("client removed", playerID)
	r.clientMutex.Lock()
	delete(r.clients, playerID)
	r.clientMutex.Unlock()
}

func (r *room) getClient(playerID string) (*client, error) {
	for _, client := range r.clients {
		if client.playerID == playerID {
			return client, nil
		}
	}

	return nil, fmt.Errorf("client with playerID %s not found", playerID)
}

func (r *room) broadcastMessage(message []byte) {
	r.clientMutex.Lock()
	defer r.clientMutex.Unlock()
	for _, client := range r.clients {
		client.messages <- message
	}
}
