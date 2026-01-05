// Package server
package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

type Server struct {
	*http.Server
}

func NewServer(queries *sqlc.Queries) *Server {
	mux := http.NewServeMux()
	registerRoutes(mux, queries)
	return &Server{
		Server: &http.Server{
			Addr:         ":8080",
			Handler:      withCORS(mux),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go s.gracefulShutdown(done)

	slog.Info("server starting", "addr", s.Addr)
	return s.ListenAndServe()
}

func (s *Server) gracefulShutdown(done <-chan os.Signal) {
	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
}

func withCORS(h http.Handler) http.Handler {
	middleware := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})
	return middleware.Handler(h)
}
