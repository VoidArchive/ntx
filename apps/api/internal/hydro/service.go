package hydro

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/gen/go/ntx/v1/ntxv1connect"
	"github.com/voidarchive/ntx/internal/company"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

const readingDays = 7

// Service serves the plant list with states computed at read time; nothing about
// a state is stored (GLOSSARY: Plant state).
type Service struct {
	ntxv1connect.UnimplementedHydroServiceHandler
	queries *sqlc.Queries
	now     func() time.Time
}

func NewService(queries *sqlc.Queries) *Service {
	return &Service{queries: queries, now: time.Now}
}

func (s *Service) ListPlants(
	ctx context.Context,
	_ *connect.Request[ntxv1.ListPlantsRequest],
) (*connect.Response[ntxv1.ListPlantsResponse], error) {
	plants, err := s.plants(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	lastRun, err := s.queries.GetLastWeatherRun(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res := &ntxv1.ListPlantsResponse{Plants: plants, Rules: Rules, LastRun: lastRun}
	states := map[string][]ntxv1.PlantState{}
	for _, p := range plants {
		states[p.Basin] = append(states[p.Basin], p.State)
	}
	for _, name := range Basins {
		res.Basins = append(res.Basins, &ntxv1.Basin{
			Name:       name,
			State:      BasinState(states[name]),
			PlantCount: int32(len(states[name])), //nolint:gosec // a few hundred plants
		})
	}
	res.Opening = MapOpening(plants, res.Basins)
	return connect.NewResponse(res), nil
}

func (s *Service) GetCompanyPlants(
	ctx context.Context,
	req *connect.Request[ntxv1.GetCompanyPlantsRequest],
) (*connect.Response[ntxv1.GetCompanyPlantsResponse], error) {
	if req.Msg.Symbol == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("symbol is required"))
	}
	plants, err := s.plants(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res := &ntxv1.GetCompanyPlantsResponse{}
	for _, p := range plants {
		if p.Symbol == req.Msg.Symbol {
			res.Plants = append(res.Plants, p)
		}
	}
	res.Opening = CompanyOpening(res.Plants)
	return connect.NewResponse(res), nil
}

func (s *Service) GetPlantReadings(
	ctx context.Context,
	req *connect.Request[ntxv1.GetPlantReadingsRequest],
) (*connect.Response[ntxv1.GetPlantReadingsResponse], error) {
	rows, err := s.queries.ListRecentReadings(ctx, sqlc.ListRecentReadingsParams{
		PlantID: req.Msg.PlantId, Limit: readingDays,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	p95, err := s.percentiles(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res := &ntxv1.GetPlantReadingsResponse{}
	for _, r := range rows {
		v, ok := p95[req.Msg.PlantId]
		res.Readings = append(res.Readings, toReading(r, v, ok))
	}
	return connect.NewResponse(res), nil
}

// plants joins every plant with its latest reading and discharge percentile and
// applies the rules. The whole table is a few hundred rows, so one pass serves
// both the map and a company's section.
func (s *Service) plants(ctx context.Context) ([]*ntxv1.Plant, error) {
	rows, err := s.queries.ListHydroPlants(ctx)
	if err != nil {
		return nil, fmt.Errorf("list plants: %w", err)
	}
	latestRows, err := s.queries.ListLatestReadings(ctx)
	if err != nil {
		return nil, fmt.Errorf("latest readings: %w", err)
	}
	latest := make(map[int64]sqlc.HydroReading, len(latestRows))
	for _, r := range latestRows {
		latest[r.PlantID] = r
	}
	p95, err := s.percentiles(ctx)
	if err != nil {
		return nil, err
	}
	today := s.now().In(nepalTime)
	out := make([]*ntxv1.Plant, 0, len(rows))
	for _, r := range rows {
		p := toPlant(r)
		if v, ok := p95[r.ID]; ok {
			p.DischargeP95 = &v
			p.Threshold = Threshold(v)
		}
		if l, ok := latest[r.ID]; ok {
			p.Latest = toReading(l, p95[r.ID], p.DischargeP95 != nil)
			p.State = p.Latest.State
			p.Stale = Stale(l.Day, today)
		}
		out = append(out, p)
	}
	return out, nil
}

// percentiles is each plant's ten-year 95th percentile of discharge, absent until it has rows.
func (s *Service) percentiles(ctx context.Context) (map[int64]float64, error) {
	floor := s.now().In(nepalTime).AddDate(-dischargeYears, 0, 0).Format(time.DateOnly)
	rows, err := s.queries.ListDischargeP95(ctx, floor)
	if err != nil {
		return nil, fmt.Errorf("discharge percentile: %w", err)
	}
	p95 := make(map[int64]float64, len(rows))
	for _, r := range rows {
		p95[r.PlantID] = r.DischargeM3s.Float64
	}
	return p95, nil
}

func toPlant(r sqlc.ListHydroPlantsRow) *ntxv1.Plant {
	p := &ntxv1.Plant{
		Id: r.ID, Symbol: r.Symbol, CompanyName: r.CompanyName.String, Plant: r.Plant, River: r.River,
		Basin: r.Basin, District: r.District, Mw: r.Mw, Lat: r.Lat, Lon: r.Lon, Status: r.Status,
		SourceUrl: r.SourceUrl, Rating: r.Rating.String, CloseDate: r.CloseDate.String,
		Period: company.Period(sqlc.Fundamental{FiscalYear: r.FiscalYear.String, Quarter: r.Quarter.String}),
	}
	if r.ClosePrice.Valid {
		close := r.ClosePrice.Float64
		p.Close = &close
		// The screener's PE: latest close over the latest quarterly EPS, loss-makers unrated.
		if r.Eps.Float64 > 0 {
			pe := close / r.Eps.Float64
			p.Pe = &pe
		}
	}
	return p
}

func toReading(r sqlc.HydroReading, p95 float64, hasP95 bool) *ntxv1.Reading {
	out := &ntxv1.Reading{
		Day: r.Day, RainSource: r.RainSource.String, DischargeSource: r.DischargeSource.String,
		State: PlantState(r, p95, hasP95),
	}
	if r.RainMm.Valid {
		out.RainMm = &r.RainMm.Float64
	}
	if r.DischargeM3s.Valid {
		out.DischargeM3S = &r.DischargeM3s.Float64
	}
	return out
}
