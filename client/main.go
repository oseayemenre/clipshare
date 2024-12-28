package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
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

func run(ctx context.Context) {
	godotenv.Load()

	clientOs := runtime.GOOS
	addr := os.Getenv("ADDR")

	if addr == " " {
		slog.Error("No address provided")
		os.Exit(1)
	}

	conn, _, err := websocket.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	defer conn.Close()

	if err != nil {
		slog.Error("Unable to connect to websocket server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var payload *Payload
	var out []byte

	for {
		select {
		case <-ctx.Done():
			slog.Info("Kill signal recieved...")
			return
		default:
			switch clientOs {
			case "windows":
				out, err = exec.Command("powershell", "-Command", "Get-Clipboard").Output()

			case "linux":
				out, err = exec.Command("xclip", "-selection", "clipboard", "-o").Output()

			case "darwin":
				out, err = exec.Command("pbpaste").Output()
			}
		}

		if err != nil {
			slog.Error("Something went wrong", slog.String("error", err.Error()))
			os.Exit(1)
		}

		payload = &Payload{Type: "broadcast_message", Message: string(out)}

		data, err := json.Marshal(payload)

		if err != nil {
			slog.Error("Unable to marshal json", slog.String("error", err.Error()))
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, data)

		if err != nil {
			slog.Error("Unable to write to websocket connection", slog.String("error", err.Error()))
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	sigctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	wg := &sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()
		run(sigctx)
	}()

	<-sigctx.Done()
	wg.Wait()
	slog.Info("Gracefully shutting down...")
}
