package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type Payload struct {
	Type    string `json:"type"`
	To      string `json:"to"`
	Message string `json:"message"`
}

func run() {
	godotenv.Load()

	clientOs := runtime.GOOS
	addr := os.Getenv("ADDR")

	if addr == " " {
		slog.Error("No address provided")
		os.Exit(1)
	}

	conn, _, err := websocket.DefaultDialer.Dial("ws://"+addr+"/ws", nil)

	if err != nil {
		slog.Error("Unable to connect to websocket server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var payload *Payload

	switch clientOs {
	case "windows":
		for {
			out, err := exec.Command("powershell", "-Command", "Get-Clipboard").Output()

			if err != nil {
				slog.Error("Something went wrong", slog.String("error", err.Error()))
				os.Exit(1)
			}

			payload = &Payload{Type: "", To: "", Message: strings.TrimSpace(string(out))}

			data, err := json.Marshal(payload)

			if err != nil {
				slog.Error("Unable to marshal json", slog.String("error", err.Error()))
				continue
			}

			err = conn.WriteMessage(websocket.TextMessage, data)

			if err != nil {
				slog.Error("Unable to write to websocket connection", slog.String("error", err.Error()))
				continue
			}

			time.Sleep(1 * time.Second)
		}

	case "linux":
		for {
			out, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output()

			if err != nil {
				slog.Error("Something went wrong", slog.String("error", err.Error()))
				os.Exit(1)
			}

			slog.Info(strings.TrimSpace(string(out)))
			time.Sleep(1 * time.Second)
		}

	case "darwin":
		for {
			out, err := exec.Command("pbpaste").Output()

			if err != nil {
				slog.Error("Something went wrong", slog.String("error", err.Error()))
				os.Exit(1)
			}

			slog.Info(strings.TrimSpace(string(out)))
			time.Sleep(1 * time.Second)
		}
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	sigctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		for {
			select {
			case <-sigctx.Done():
				slog.Info("Kill signal recieved...")
				return
			default:
				run()
			}
		}
	}()

	<-sigctx.Done()
	slog.Info("Gracefully shutting down...")
}
