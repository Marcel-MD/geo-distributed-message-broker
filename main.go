package main

import (
	"geo-distributed-message-broker/api"
	"geo-distributed-message-broker/config"
	"geo-distributed-message-broker/data"
	"geo-distributed-message-broker/domain"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Config
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err.Error())
		return
	}

	// Database
	db, err := data.NewDB()
	if err != nil {
		slog.Error("Failed to create database connection", "error", err.Error())
		return
	}

	repo := data.NewRepository(db)
	broker := domain.NewBroker(repo)

	// GRPC Server
	grpcSrv, listener, err := api.NewServer(cfg, broker)
	if err != nil {
		slog.Error("Failed to create GRPC server", "error", err.Error())
		return
	}

	slog.Info("Starting GRPC server")
	go func() {
		if err := grpcSrv.Serve(listener); err != nil {
			slog.Error("Failed to start gRPC server", "error", err.Error())
			return
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)
	<-quit
	slog.Warn("Shutting down server")

	// Shutdown GRPC server
	grpcSrv.GracefulStop()

	// Close DB connection
	if err := data.CloseDB(db); err != nil {
		slog.Error("Failed to close database connection", "error", err.Error())
		return
	}
}
