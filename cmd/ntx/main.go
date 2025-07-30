package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anish/ntx/internal/config"
	"github.com/anish/ntx/internal/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := logger.New(cfg.Logger)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize and run application
	app := NewApp(cfg, logger)
	if err := app.Run(ctx); err != nil {
		logger.Error("Application failed", "error", err)
		os.Exit(1)
	}
}