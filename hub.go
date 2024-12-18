package main

import "sync"

type hub struct {
	clients map[*client]bool
	mutex   *sync.Mutex
}

func NewHub() *hub {
	return &hub{clients: make(map[*client]bool), mutex: &sync.Mutex{}}
}

func (h *hub) registerClient(client *client) {
	h.mutex.Lock()
	h.clients[client] = true
	defer h.mutex.Unlock()
}

func (h *hub) unregisterClient(client *client) {
	h.mutex.Lock()

	if _, ok := h.clients[client]; ok {
		client.conn.Close()
		delete(h.clients, client)
	}

	defer h.mutex.Unlock()
}
