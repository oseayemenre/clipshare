package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	sigctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer cancel()

	newServer := NewServer(getPrivateIp())
	svr := newServer.buildHTTPServer()

	go func() {
		slog.Info("Server is listening on " + newServer.addr)
		if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server encountered an error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-sigctx.Done()
	slog.Info("Shutdown signal recieved...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := svr.Shutdown(ctx); err != nil {
		slog.Error("Unable to perform clean shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("Server gracefully shutting down...")
}
