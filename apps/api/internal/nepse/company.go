package nepse

import (
	"context"
	"fmt"
)

type Company struct {
	ID             int64
	Name           string
	Symbol         string
	Status         string
	Email          string
	Website        string
	Sector         string
	InstrumentType string
}

func (c *Client) Companies(ctx context.Context) ([]Company, error) {
	companyList, err := c.api.Companies(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch companies: %w", err)
	}
	var companies []Company
	for _, co := range companyList {
		if co.InstrumentType != "Equity" || co.Status != "A" {
			continue
		}
		companies = append(companies, Company{
			ID:             int64(co.ID),
			Name:           co.CompanyName,
			Symbol:         co.Symbol,
			Status:         co.Status,
			Email:          co.CompanyEmail,
			Website:        co.Website,
			Sector:         co.SectorName,
			InstrumentType: co.InstrumentType,
		})
	}
	return companies, nil
}
