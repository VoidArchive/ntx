package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"ntx/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Shutdown signal received, shutting down gracefully...")
		cancel()
	}()

	// Initialize and run the application
	if err := run(ctx); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Initialize the application
	application, err := app.New()
	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	// Start the application
	if err := application.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	// Wait for context cancellation (shutdown signal)
	<-ctx.Done()

	// Graceful shutdown
	if err := application.Stop(); err != nil {
		return fmt.Errorf("failed to stop application gracefully: %w", err)
	}

	return nil
}