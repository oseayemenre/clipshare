package main

import "log/slog"

type hub struct {
	clients    map[string]*client
	register   chan *client
	unregister chan *client
	broadcast  chan []byte
}

func NewHub() *hub {
	return &hub{
		clients:    make(map[string]*client),
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan []byte),
	}
}

func (h *hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.id] = client
			slog.Info("Client registered", slog.String("id", client.id))
		case client := <-h.unregister:
			client.conn.Close()
			delete(h.clients, client.id)
			slog.Info("Client unregistered", slog.String("id", client.id))
		case message := <-h.broadcast:
			for id := range h.clients {
				client := h.clients[id]
				client.egress <- message
			}
		}
	}
}
