package server

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/auth"
	"github.com/voidarchive/ntx/internal/company"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/portfolio"
	"github.com/voidarchive/ntx/internal/price"
)

func registerRoutes(mux *http.ServeMux, queries *sqlc.Queries) {
	// Create auth service (needed for both login and middleware)
	authService := auth.NewAuthService(queries)
	authInterceptor := auth.NewAuthInterceptor(authService)

	// Create interceptors slice
	interceptors := connect.WithInterceptors(authInterceptor)

	// Public services (no auth required, but interceptor skips these)
	companyPath, companyHandler := ntxv1connect.NewCompanyServiceHandler(
		company.NewCompanyService(queries),
		interceptors,
	)
	mux.Handle(companyPath, companyHandler)

	pricePath, priceHandler := ntxv1connect.NewPriceServiceHandler(
		price.NewPriceService(queries),
		interceptors,
	)
	mux.Handle(pricePath, priceHandler)

	// Auth service (login is public)
	authPath, authHandler := ntxv1connect.NewAuthServiceHandler(
		authService,
		interceptors,
	)
	mux.Handle(authPath, authHandler)

	// Protected services
	portfolioPath, portfolioHandler := ntxv1connect.NewPortfolioServiceHandler(
		portfolio.NewPortfolioService(queries),
		interceptors,
	)
	mux.Handle(portfolioPath, portfolioHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
