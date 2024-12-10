package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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

func (s *server) handleConnections(w *websocket.Conn) {}

func (s *server) run() error {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}

		ws, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Println(err)
		}
	})

	log.Printf("Server is listening on %s", s.addr)
	return http.ListenAndServe(s.addr, r)
}
