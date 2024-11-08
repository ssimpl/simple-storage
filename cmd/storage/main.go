package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ssimpl/simple-storage/internal/storage/service"
	"github.com/ssimpl/simple-storage/internal/storage/transport/grpc"
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

	slog.Info("Starting Storage server", "addr", cfg.Addr)

	objectStorage := service.NewObjectStorage(cfg.StoragePath)
	storageSrv := grpc.NewStorageServer(objectStorage)

	server := grpc.NewServer(cfg.Addr, storageSrv)

	go func() {
		if err := server.Start(); err != nil {
			slog.Error("Start Storage server error", "err", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()

	slog.Info("Shutting down Storage server")

	server.Stop()

	slog.Info("Storage server stopped")

	return nil
}
