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
	"github.com/voidarchive/ntx/internal/hydro"
	"github.com/voidarchive/ntx/internal/nepse"
	"github.com/voidarchive/ntx/internal/server"
	"github.com/voidarchive/ntx/internal/story"
	"github.com/voidarchive/ntx/internal/worker"
)

func main() {
	// Configure slog to write to stdout (Railway treats stderr as errors)
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

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
	companies        bool
	fundamentals     bool
	prices           bool
	ownership        bool
	corporateActions bool
	ratings          bool
	weather          bool
}

func runBackfillCmd() {
	fs := flag.NewFlagSet("backfill", flag.ExitOnError)
	opts := backfillOptions{}
	fs.BoolVar(&opts.companies, "companies", false, "sync companies")
	fs.BoolVar(&opts.fundamentals, "fundamentals", false, "sync fundamentals")
	fs.BoolVar(&opts.prices, "prices", false, "sync price history")
	fs.BoolVar(&opts.ownership, "ownership", false, "sync ownership data")
	fs.BoolVar(&opts.corporateActions, "corporate-actions", false, "sync corporate actions")
	fs.BoolVar(&opts.ratings, "ratings", false, "store every company's rating from the story engine")
	fs.BoolVar(&opts.weather, "weather", false, "fetch hydro plant readings now instead of at 06:00")
	_ = fs.Parse(os.Args[2:])

	// If no flags specified, sync everything
	if !opts.companies && !opts.fundamentals && !opts.prices && !opts.ownership && !opts.corporateActions &&
		!opts.ratings && !opts.weather {
		opts.companies = true
		opts.fundamentals = true
		opts.prices = true
		opts.ownership = true
		opts.corporateActions = true
		opts.ratings = true
	}

	db, queries, client := setup()
	defer db.Close()
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := runBackfill(ctx, worker.New(client, queries), opts); err != nil {
		slog.Error("backfill failed", "error", err)
		os.Exit(1)
	}
}

func runServer() {
	db, queries, client := setup()
	defer db.Close()

	w := worker.New(client, queries)
	// The plant table is the CSV; every start re-reads it so a CSV edit plus a deploy is the workflow (ADR 0004).
	if err := hydro.SyncPlants(context.Background(), queries, plantsPath()); err != nil {
		slog.Error("hydro plants sync failed", "error", err)
		os.Exit(1)
	}
	sched, err := worker.NewScheduler(w)
	if err != nil {
		slog.Error("scheduler init failed", "error", err)
		os.Exit(1)
	}
	go func() {
		_ = sched.Start(context.Background())
	}()
	// A fresh deploy has no ratings until the post-close job runs, a deploy that
	// changed the rules has only stale ones, and a sync that stopped part-way
	// left some; fill them now so the screener is not blank or mixed for a day.
	go func() {
		ctx := context.Background()
		n, err := queries.CountUnratedCompanies(ctx, story.RatingVersion)
		if err != nil {
			slog.Error("count ratings", "error", err)
			return
		}
		if n == 0 {
			return
		}
		if err := w.SyncRatings(ctx); err != nil {
			slog.Error("initial ratings failed", "error", err)
		}
	}()

	srv := server.NewServer(db, queries)
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

// plantsPath is data/hydro_plants.csv: copied beside the binary in the image, pointed at by NTX_PLANTS_CSV in dev.
func plantsPath() string {
	if p := os.Getenv("NTX_PLANTS_CSV"); p != "" {
		return p
	}
	return "data/hydro_plants.csv"
}
