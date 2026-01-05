package main

import (
	"log/slog"
	"os"

	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/server"
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
	srv := server.NewServer(queries)

	if err := srv.Start(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
