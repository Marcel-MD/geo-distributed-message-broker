package main

import (
	"fmt"
	"geo-distributed-message-broker/api"
	"geo-distributed-message-broker/config"
	"geo-distributed-message-broker/data"
	"geo-distributed-message-broker/services"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
)

func main() {
	// Logger
	configureLogger()

	// Config
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err.Error())
		return
	}

	// Database
	db, err := data.NewDB(cfg)
	if err != nil {
		slog.Error("Failed to create database connection", "error", err.Error())
		return
	}

	repo := data.NewRepository(db)
	broker := services.NewBrokerService(repo)
	consensus := services.NewConsensusService(cfg, broker)

	// Broker Server
	brokerSrv, brokerListener, err := api.NewBrokerServer(cfg, broker, consensus)
	if err != nil {
		slog.Error("Failed to create broker server", "error", err.Error())
		return
	}

	slog.Info(fmt.Sprintf("Starting broker server on port %s 🚀", cfg.BrokerPort))
	go func() {
		if err := brokerSrv.Serve(brokerListener); err != nil {
			slog.Error("Failed to start broker server", "error", err.Error())
			return
		}
	}()

	// Node Server
	nodeSrv, nodeListener, err := api.NewNodeServer(cfg, consensus)
	if err != nil {
		slog.Error("Failed to create node server", "error", err.Error())
		return
	}

	slog.Info(fmt.Sprintf("Starting node server on port %s 🚀", cfg.NodePort))
	go func() {
		if err := nodeSrv.Serve(nodeListener); err != nil {
			slog.Error("Failed to start node server", "error", err.Error())
			return
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)
	<-quit
	slog.Warn("Shutting down server ⛔")

	// Shutdown broker server
	brokerSrv.GracefulStop()

	// Shutdown node server
	nodeSrv.GracefulStop()

	// Close DB connection
	if err := data.CloseDB(db); err != nil {
		slog.Error("Failed to close database connection", "error", err.Error())
		return
	}

	slog.Info("Server successful shutdown ✅")
}

func configureLogger() {
	var handler slog.Handler = tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.RFC3339,
	})

	l := slog.New(handler)
	slog.SetDefault(l)
}
