package ws

import (
	"log"
	"sync"
)

type room struct {
	clients     map[*client]struct{}
	register    chan *client
	unregister  chan *client
	broadcast   chan []byte
	clientMutex sync.Mutex
	quit        chan bool
}

func NewRoom() *room {
	return &room{
		clients:    make(map[*client]struct{}),
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
			r.removeClient(client)

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
	r.clients[client] = struct{}{}
	r.clientMutex.Unlock()
}

func (r *room) removeClient(client *client) {
	log.Println("client removed", client)
	r.clientMutex.Lock()
	delete(r.clients, client)
	r.clientMutex.Unlock()
}

func (r *room) broadcastMessage(message []byte) {
	r.clientMutex.Lock()
	defer r.clientMutex.Unlock()
	for client := range r.clients {
		client.messages <- message
	}
}
