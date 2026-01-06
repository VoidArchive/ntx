package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/nepse"
	"github.com/voidarchive/ntx/internal/server"
	"github.com/voidarchive/ntx/internal/worker"
)

func main() {
	dbPath := database.DefaultPath()
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

	nepseClient, err := nepse.NewClient()
	if err != nil {
		slog.Error("nepse client", "error", err)
		os.Exit(1)
	}

	// HACK: need to fix and move this to correct file
	w := worker.New(nepseClient, queries)
	sched, err := worker.NewScheduler(w)
	if err != nil {
		slog.Error("scheduler init failed", "error", err)
		os.Exit(1)
	}
	// if err := w.SyncCompanies(context.Background()); err != nil {
	// 	slog.Error("manual sync failed", "error", err)
	// 	os.Exit(1)
	// }
	go func() {
		_ = sched.Start(context.Background())
	}()

	// NOTE: Start
	srv := server.NewServer(queries)
	if err := srv.Start(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
