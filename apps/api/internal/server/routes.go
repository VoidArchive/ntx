package server

import (
	"net/http"

	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/company"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/market"
	"github.com/voidarchive/ntx/internal/price"
	"github.com/voidarchive/ntx/internal/sector"
	"github.com/voidarchive/ntx/internal/story"
)

// registerRoutes is the only place services are mounted; NTX has no accounts (ADR 0003).
func registerRoutes(mux *http.ServeMux, queries *sqlc.Queries) {
	mux.Handle(ntxv1connect.NewCompanyServiceHandler(company.NewCompanyService(queries)))
	mux.Handle(ntxv1connect.NewPriceServiceHandler(price.NewPriceService(queries)))
	mux.Handle(ntxv1connect.NewStoryServiceHandler(story.NewService(queries)))
	mux.Handle(ntxv1connect.NewMarketServiceHandler(market.NewMarketService(queries)))
	mux.Handle(ntxv1connect.NewSectorServiceHandler(sector.NewService(queries)))

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
