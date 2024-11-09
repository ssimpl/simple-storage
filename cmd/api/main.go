package main

import (
	"context"
	"log"
	"log/slog"
	nh "net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ssimpl/simple-storage/internal/api/infrastructure/db/pg"
	"github.com/ssimpl/simple-storage/internal/api/infrastructure/storage"
	"github.com/ssimpl/simple-storage/internal/api/service"
	"github.com/ssimpl/simple-storage/internal/api/transport/http"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := newConfig()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting API server", "addr", cfg.Addr)

	storageClient := storage.NewClient()
	metaRepo, err := pg.NewDB(pg.Config{
		Addr:     cfg.PG.Addr,
		Database: cfg.PG.Database,
		User:     cfg.PG.User,
		Password: cfg.PG.Password,
		SQLDebug: cfg.PG.SQLDebug,
		AppName:  cfg.PG.AppName,
		Timeout:  cfg.PG.Timeout,
	})
	if err != nil {
		return err
	}

	objectManager := service.NewObjectManager(storageClient, metaRepo, cfg.FileFragments)
	handler := http.NewHandler(objectManager, cfg.FileSizeLimit)

	mux := nh.NewServeMux()
	mux.HandleFunc("/", handler.ServeHTTP)

	server := http.NewServer(cfg.Addr, mux)

	go func() {
		if err := server.Start(); err != nil {
			slog.Error("Start API server error", "err", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()

	slog.Info("Shutting down API server")

	if err := server.Stop(); err != nil {
		slog.Error("Stop API server error", "err", err)
	}

	slog.Info("API server stopped")

	return nil
}
