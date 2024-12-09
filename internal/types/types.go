package types

import (
	"golang.org/x/net/websocket"
)

type Server struct {
	Addr string
	Conn map[websocket.Conn]bool
}
