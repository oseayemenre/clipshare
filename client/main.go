package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type Payload struct {
	Type    string `json:"type"`
	To      string `json:"to"`
	Message string `json:"message"`
}

func run(ctx context.Context) (string, error) {
	godotenv.Load()

	sigctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	clientOs := runtime.GOOS
	addr := os.Getenv("ADDR")

	if addr == " " {
		slog.Error("No address provided")
		os.Exit(1)
	}
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	conn.SetPongHandler(func(msg string) error {
		slog.Info("PONG")
		return conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	})

	if err != nil {
		return "Unable to connect to websocket server", err
	}

	for {
		select {
		case <-sigctx.Done():
			slog.Info("Kill signal recieved...")
			return "", nil
		case <-ticker.C:

			go func() {
				for {
					_, _, err := conn.ReadMessage()
					if err != nil {
						slog.Error("Error reading message", slog.String("error", err.Error()))
						return
					}
				}
			}()
			var out []byte

			switch clientOs {
			case "windows":
				out, err = exec.Command("powershell", "-Command", "Get-Clipboard").Output()

			case "linux":
				out, err = exec.Command("xclip", "-selection", "clipboard", "-o").Output()

			case "darwin":
				out, err = exec.Command("pbpaste").Output()
			}

			data, err := json.Marshal(&Payload{Type: "broadcast_message", Message: string(out)})

			if err != nil {
				return "Unable to marshal data", err
			}

			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				return "Connection closed", err
			}
		}

	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	ctx := context.Background()

	if data, err := run(ctx); err != nil {
		slog.Error(data, slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Gracefully shutting down...")
}
