package main

import (
	"log/slog"

	"github.com/gorilla/websocket"
)

type client struct {
	conn *websocket.Conn
	*hub
	egress chan []byte
}

func NewClient(ws *websocket.Conn, hub *hub) *client {
	return &client{conn: ws, hub: hub}
}

func (c *client) readLoop() {
	defer c.hub.unregisterClient(c)

	_, message, err := c.conn.ReadMessage()

	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			slog.Error(err.Error())
		}
		return
	}

	slog.Debug(string(message))

	for client := range c.hub.clients {
		client.egress <- message
	}
}

func (c *client) writeLoop() {
	defer c.hub.unregisterClient(c)

	for {
		message, ok := <-c.egress

		if !ok {
			if err := c.conn.WriteMessage(websocket.CloseMessage, []byte("connection closed")); err != nil {
				slog.Error("connection closed")
			}
			return
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			slog.Error("Something went wrong")
			return
		}

		slog.Info("Message sent")
	}
}
