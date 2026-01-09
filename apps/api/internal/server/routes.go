package server

import (
	"net/http"

	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/company"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/price"
)

func registerRoutes(mux *http.ServeMux, queries *sqlc.Queries) {
	companyPath, companyHandler := ntxv1connect.NewCompanyServiceHandler(
		company.NewCompanyService(queries),
	)
	mux.Handle(companyPath, companyHandler)

	pricePath, priceHandler := ntxv1connect.NewPriceServiceHandler(
		price.NewPriceService(queries),
	)
	mux.Handle(pricePath, priceHandler)
}
