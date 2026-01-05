package server

import (
	"net/http"

	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/company"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

func registerRoutes(mux *http.ServeMux, queries *sqlc.Queries) {
	companyPath, companyHandler := ntxv1connect.NewCompanyServiceHandler(
		company.NewCompanyService(queries),
	)
	mux.Handle(companyPath, companyHandler)
}
