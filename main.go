package main

import (
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	svr := NewServer(getPrivateIp())

	if err := svr.run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
