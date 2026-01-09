package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/nepse"
	"github.com/voidarchive/ntx/internal/server"
	"github.com/voidarchive/ntx/internal/worker"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "backfill":
			runBackfillCmd()
			return
		case "serve":
			runServer()
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
			fmt.Fprintln(os.Stderr, "usage: ntx [backfill|serve]")
			os.Exit(1)
		}
	}

	// Default: run server
	runServer()
}

type backfillOptions struct {
	companies    bool
	fundamentals bool
	prices       bool
}

func runBackfillCmd() {
	fs := flag.NewFlagSet("backfill", flag.ExitOnError)
	opts := backfillOptions{}
	fs.BoolVar(&opts.companies, "companies", false, "sync companies")
	fs.BoolVar(&opts.fundamentals, "fundamentals", false, "sync fundamentals")
	fs.BoolVar(&opts.prices, "prices", false, "sync price history")
	fs.Parse(os.Args[2:])

	// If no flags specified, sync everything
	if !opts.companies && !opts.fundamentals && !opts.prices {
		opts.companies = true
		opts.fundamentals = true
		opts.prices = true
	}

	db, queries, client := setup()
	defer db.Close()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := runBackfill(ctx, queries, client, opts); err != nil {
		slog.Error("backfill failed", "error", err)
		os.Exit(1)
	}
}

func runServer() {
	db, queries, client := setup()
	defer db.Close()

	w := worker.New(client, queries)
	sched, err := worker.NewScheduler(w)
	if err != nil {
		slog.Error("scheduler init failed", "error", err)
		os.Exit(1)
	}
	go func() {
		_ = sched.Start(context.Background())
	}()

	srv := server.NewServer(queries)
	if err := srv.Start(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func setup() (*sql.DB, *sqlc.Queries, *nepse.Client) {
	dbPath := database.DefaultPath()
	db, err := database.OpenDB(dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

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

	return db, queries, nepseClient
}
