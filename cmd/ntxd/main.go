package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/voidarchive/ntx/cmd/ntxd/handlers"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/market"
	"github.com/voidarchive/ntx/internal/nepse"
	"github.com/voidarchive/ntx/internal/worker"
)

func main() {
	dbPath := database.DefaultServerPath()
	db, err := database.OpenDB(dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("database opened", "path", dbPath)

	if err := database.AutoMigrate(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("database initialized")

	queries := sqlc.New(db)
	mkt := market.New(queries)

	// Create NEPSE client
	nepseClient, err := nepse.NewClient()
	if err != nil {
		slog.Error("failed to create nepse client", "error", err)
		os.Exit(1)
	}
	defer func() { _ = nepseClient.Close() }()

	// Create worker
	w := worker.New(nepseClient, queries, mkt)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker in background
	go func() {
		if err := w.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("worker error", "error", err)
		}
	}()

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Register RPC handlers
	companyPath, companyHandler := ntxv1connect.NewCompanyServiceHandler(
		handlers.NewCompanyService(queries),
	)
	mux.Handle(companyPath, companyHandler)

	pricePath, priceHandler := ntxv1connect.NewPriceServiceHandler(
		handlers.NewPriceService(queries),
	)
	mux.Handle(pricePath, priceHandler)

	marketPath, marketHandler := ntxv1connect.NewMarketServiceHandler(
		handlers.NewMarketService(queries, mkt, nepseClient),
	)
	mux.Handle(marketPath, marketHandler)

	screenerPath, screenerHandler := ntxv1connect.NewScreenerServiceHandler(
		handlers.NewScreenerService(queries),
	)
	mux.Handle(screenerPath, screenerHandler)

	addr := ":8080"
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-done
		slog.Info("shutting down")

		// Cancel worker context first
		cancel()

		// Then shutdown HTTP server
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}()

	slog.Info("server starting", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
