// Package server
package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Server{
		Server: &http.Server{
			Addr:         ":" + port,
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
	origins := getCORSOrigins()
	middleware := cors.New(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})
	return middleware.Handler(h)
}

func getCORSOrigins() []string {
	env := os.Getenv("CORS_ORIGINS")
	if env == "" {
		return []string{"http://localhost:5173"}
	}

	var origins []string
	for o := range strings.SplitSeq(env, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
