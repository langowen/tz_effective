package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"tz_effective/deploy/config"
	"tz_effective/internal/adaper/storage/postgres"
	"tz_effective/internal/ports/http/public"
	"tz_effective/internal/service"
)

func main() {

	cfg := config.NewConfig()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())

	pgStorage, err := postgres.New(ctx, cfg)
	if err != nil {
		log.Fatalln("Failed to initialize PostgresSQL storage", "error", err)
	}

	serviceRate := service.NewService(pgStorage, cfg)

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	logger.With(
		"Config params", cfg,
		"go_version", runtime.Version(),
	).Info("starting server")

	serverDone := public.StartServer(ctx, serviceRate, cfg)

	logger.Info("server started")

	<-done
	cancel()
	logger.Info("stopping server")

	<-serverDone
	logger.Info("server stopped")

}
