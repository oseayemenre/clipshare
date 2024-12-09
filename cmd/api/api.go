package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/oseayemenre/clip_share/internal/types"
	"golang.org/x/net/websocket"
)

type server struct {
	*types.Server
}

func NewServer(addr string) *server {
	return &server{
		Server: &types.Server{
			Addr: addr,
			Conn: make(map[websocket.Conn]bool),
		},
	}
}

func (s *server) handleConnections(w *websocket.Conn) {}

func (s *server) run() error {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	r.Handle("/ws", websocket.Handler(s.handleConnections))

	log.Printf("Server is listening on %s", s.Server.Addr)
	return http.ListenAndServe(s.Server.Addr, r)
}
