package main

import (
	"log/slog"
	"sync"
)

type hub struct {
	clients    map[string]*client
	register   chan *client
	unregister chan *client
	broadcast  chan []byte
	mu         *sync.Mutex
}

func NewHub() *hub {
	return &hub{
		clients:    make(map[string]*client),
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan []byte),
		mu:         &sync.Mutex{},
	}
}

func (h *hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.id] = client
			h.mu.Unlock()
			slog.Info("Client registered", slog.String("id", client.id))
		case client := <-h.unregister:
			client.conn.Close()
			h.mu.Lock()
			delete(h.clients, client.id)
			h.mu.Unlock()
			slog.Info("Client unregistered", slog.String("id", client.id))
		case message := <-h.broadcast:
			for id := range h.clients {
				client := h.clients[id]
				client.egress <- message
			}
		}
	}
}
