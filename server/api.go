package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type server struct {
	addr string
}

func NewServer(addr string) *server {
	return &server{
		addr,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *server) buildHTTPServer() *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	h := NewHub()
	go h.run()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is up and running"))
	})

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			slog.Error("Unable to upgrade connection", slog.String("error", err.Error()))
		}

		client := NewClient(conn, uuid.New().String(), h)
		h.register <- client

		go client.readLoop()
		go client.writeLoop()
	})

	server := &http.Server{
		Addr:    s.addr,
		Handler: r,
	}

	return server
}
