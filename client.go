package main

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

type client struct {
	id   string
	conn *websocket.Conn
	*hub
	egress chan []byte
}

type Event struct {
	Type    string `json:"type"`
	To      string `json:"to"`
	Message string `json:"message"`
}

var (
	pongWait     = 10 * time.Second
	pingInterval = pongWait * 9 / 10
)

func NewClient(ws *websocket.Conn, id string, hub *hub) *client {
	return &client{id: id, conn: ws, hub: hub, egress: make(chan []byte)}
}

func (c *client) readLoop() {
	defer func() {
		c.hub.unregister <- c
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(pongMsg string) error {
		slog.Debug("PONG!")
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	c.conn.SetReadLimit(512)

	for {
		_, message, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error(err.Error())
			}
			return
		}

		var event *Event

		if err := json.Unmarshal(message, &event); err != nil {
			slog.Error("Invalid message", slog.String("error", err.Error()))
			continue
		}

		switch event.Type {
		case "private_message":
			c.egress <- message

		case "broadcast_message":
			c.hub.broadcast <- message
		default:
			continue
		}

	}
}

func (c *client) writeLoop() {
	defer func() {
		c.hub.unregister <- c
	}()

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
			slog.Info("PING!")
		}
	}
}
