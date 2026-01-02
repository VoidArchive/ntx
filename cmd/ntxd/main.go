package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	connect "connectrpc.com/connect"

	v1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	ntxv1connect "github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/portfolio"
)

type PortfolioServer struct {
	service *portfolio.Service
}

const maxCSVSize = 10 * 1024 * 1024 // 10MB

func (s *PortfolioServer) Import(
	ctx context.Context,
	req *connect.Request[v1.ImportRequest],
) (*connect.Response[v1.ImportResponse], error) {
	if len(req.Msg.CsvData) > maxCSVSize {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("CSV exceeds maximum size of %d bytes", maxCSVSize))
	}

	result, err := s.service.ImportCSV(ctx, req.Msg.CsvData)
	if err != nil {
		slog.Error("import failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.ImportResponse{
		Imported: int32(result.Imported),
		Skipped:  int32(result.Skipped),
		Errors:   result.Errors,
	}), nil
}

func (s *PortfolioServer) ListHoldings(
	ctx context.Context,
	req *connect.Request[v1.ListHoldingsRequest],
) (*connect.Response[v1.ListHoldingsResponse], error) {
	holdings, err := s.service.ListHoldings(ctx)
	if err != nil {
		slog.Error("list holdings failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.ListHoldingsResponse{
		Holdings: holdings,
	}), nil
}

func (s *PortfolioServer) GetHolding(
	ctx context.Context,
	req *connect.Request[v1.GetHoldingRequest],
) (*connect.Response[v1.GetHoldingResponse], error) {
	if req.Msg.Symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("symbol required"))
	}

	holding, err := s.service.GetHolding(ctx, req.Msg.Symbol)
	if err != nil {
		slog.Error("get holding failed", "symbol", req.Msg.Symbol, "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.GetHoldingResponse{
		Holding: holding,
	}), nil
}

func (s *PortfolioServer) Summary(
	ctx context.Context,
	req *connect.Request[v1.SummaryRequest],
) (*connect.Response[v1.SummaryResponse], error) {
	summary, err := s.service.Summary(ctx)
	if err != nil {
		slog.Error("summary failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.SummaryResponse{
		Summary: summary,
	}), nil
}

func (s *PortfolioServer) ListTransactions(
	ctx context.Context,
	req *connect.Request[v1.ListTransactionsRequest],
) (*connect.Response[v1.ListTransactionsResponse], error) {
	limit := int32(100)
	offset := int32(0)
	symbol := ""
	txType := v1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED

	if req.Msg.Limit > 0 {
		limit = req.Msg.Limit
	}
	if req.Msg.Offset > 0 {
		offset = req.Msg.Offset
	}
	if req.Msg.Symbol != "" {
		symbol = req.Msg.Symbol
	}
	if req.Msg.Type != v1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED {
		txType = req.Msg.Type
	}

	transactions, total, err := s.service.ListTransactions(ctx, symbol, txType, limit, offset)
	if err != nil {
		slog.Error("list transactions failed", "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&v1.ListTransactionsResponse{
		Transactions: transactions,
		Total:        total,
	}), nil
}

func main() {
	db, err := database.OpenDB()
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.AutoMigrate(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("database initialized")

	path, handler := ntxv1connect.NewPortfolioServiceHandler(&PortfolioServer{
		service: portfolio.NewService(db),
	})

	mux := http.NewServeMux()
	mux.Handle(path, handler)

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
		slog.Info("shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}()

	slog.Info("server starting", "addr", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
