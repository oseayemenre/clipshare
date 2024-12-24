package main

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

type client struct {
	conn *websocket.Conn
	*hub
	egress chan []byte
}

var (
	pongWait     = 10 * time.Second
	pingInterval = pongWait * 9 / 10
)

func NewClient(ws *websocket.Conn, hub *hub) *client {
	return &client{conn: ws, hub: hub}
}

func (c *client) readLoop() {
	defer c.hub.unregisterClient(c)

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(pongMsg string) error {
		slog.Debug("PONG!")
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	c.conn.SetReadLimit(512)

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

	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case message, ok := <-c.egress:
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
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte("")); err != nil {
				slog.Error("Ping error", slog.String("error", err.Error()))
			}
			slog.Debug("PING!")
		}
	}
}
