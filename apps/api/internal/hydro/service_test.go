package hydro

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// fixture seeds three plants in two basins: Chilime (Gandaki) with heavy rain
// today, Sanjen (Gandaki) with twenty days of discharge and a spike today, and
// Mai (Kankai) whose last reading is five days old. Chilime's company has a
// close, an EPS and a rating; Sanjen's company has none.
func fixture(t *testing.T) *Service {
	t.Helper()
	ctx := context.Background()
	q := database.OpenTestDB(t)
	svc := NewService(q)
	svc.now = func() time.Time { return today }

	plants := []Plant{
		{
			Symbol:    "CHCL",
			Plant:     "Chilime",
			River:     "Chilime",
			Basin:     "Gandaki",
			District:  "Rasuwa",
			Mw:        22,
			Lat:       28.17,
			Lon:       85.32,
			Status:    "operating",
			SourceUrl: "https://doed.gov.np",
		},
		{
			Symbol:    "SJCL",
			Plant:     "Sanjen",
			River:     "Sanjen",
			Basin:     "Gandaki",
			District:  "Rasuwa",
			Mw:        42.5,
			Lat:       28.2,
			Lon:       85.29,
			Status:    "under_construction",
			SourceUrl: "https://doed.gov.np",
		},
		{
			Symbol:    "SHPC",
			Plant:     "Mai",
			River:     "Mai Khola",
			Basin:     "Kankai",
			District:  "Ilam",
			Mw:        22,
			Lat:       26.81,
			Lon:       87.89,
			Status:    "operating",
			SourceUrl: "https://doed.gov.np",
		},
	}
	for _, p := range plants {
		if err := q.UpsertHydroPlant(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	if err := q.UpsertCompany(ctx, sqlc.UpsertCompanyParams{
		ID:             1,
		Name:           "Chilime Hydropower Company Limited",
		Symbol:         "CHCL",
		Status:         "A",
		Sector:         "Hydro Power",
		InstrumentType: "Equity",
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.UpsertPrice(ctx, sqlc.UpsertPriceParams{
		CompanyID: 1, BusinessDate: "2026-11-04", ClosePrice: sql.NullFloat64{Float64: 500, Valid: true},
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.UpsertFundamental(ctx, sqlc.UpsertFundamentalParams{
		CompanyID: 1, FiscalYear: "2026-2027", Quarter: "First Quarter", Eps: sql.NullFloat64{Float64: 25, Valid: true},
		FiscalYearNp: sql.NullString{String: "2083/84", Valid: true},
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.UpsertCompanyRating(
		ctx,
		sqlc.UpsertCompanyRatingParams{CompanyID: 1, Letter: "B", Holds: 5, Evaluated: 8},
	); err != nil {
		t.Fatal(err)
	}

	rain := func(plant int64, d string, mm float64) {
		if err := q.UpsertRainReading(ctx, sqlc.UpsertRainReadingParams{
			PlantID: plant, Day: d, RainMm: nullFloat(mm), RainSource: nullString(RainSource),
		}); err != nil {
			t.Fatal(err)
		}
	}
	discharge := func(plant int64, d string, m3s float64) {
		if err := q.UpsertDischargeReading(ctx, sqlc.UpsertDischargeReadingParams{
			PlantID: plant, Day: d, DischargeM3s: nullFloat(m3s), DischargeSource: nullString(DischargeSource),
		}); err != nil {
			t.Fatal(err)
		}
	}
	rain(1, day(0), 72)
	discharge(1, day(0), 10)
	// Twenty days 1..20 m3/s then a spike: p95 of 21 values is the 20th ranked, 20; 55 > 20.
	for i := 1; i <= 20; i++ {
		discharge(2, day(-i), float64(i))
		rain(2, day(-i), 1)
	}
	discharge(2, day(0), 55)
	rain(2, day(0), 0)
	rain(3, day(-5), 2)
	discharge(3, day(-5), 7)
	return svc
}

func TestListPlantsAppliesRulesAndBasins(t *testing.T) {
	svc := fixture(t)
	res, err := svc.ListPlants(context.Background(), connect.NewRequest(&ntxv1.ListPlantsRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, p := range res.Msg.Plants {
		got = append(
			got,
			fmt.Sprintf(
				"%s %s %v stale=%v p95=%v latest=%s",
				p.Symbol,
				p.Basin,
				p.State,
				p.Stale,
				p.DischargeP95 != nil,
				p.Latest.GetDay(),
			),
		)
	}
	want := []string{
		"CHCL Gandaki PLANT_STATE_HEAVY_RAIN stale=false p95=true latest=" + day(0),
		"SJCL Gandaki PLANT_STATE_HIGH_DISCHARGE stale=false p95=true latest=" + day(0),
		"SHPC Kankai PLANT_STATE_NORMAL stale=true p95=true latest=" + day(-5),
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if p := res.Msg.Plants[1]; *p.DischargeP95 != 20 {
		t.Fatalf("SJCL p95 %v, want 20", *p.DischargeP95)
	}
	if len(res.Msg.Rules) != 2 || res.Msg.LastRun == "" {
		t.Fatalf("rules %d, last run %q", len(res.Msg.Rules), res.Msg.LastRun)
	}
	wantOpening := "3 plants of 3 listed companies on 3 rivers. Basins on " + day(
		0,
	) + ": Gandaki high discharge, Kankai normal."
	if res.Msg.Opening != wantOpening {
		t.Fatalf("opening %q", res.Msg.Opening)
	}
}

func TestListPlantsJoinsCompanyFiguresAndBasins(t *testing.T) {
	svc := fixture(t)
	res, err := svc.ListPlants(context.Background(), connect.NewRequest(&ntxv1.ListPlantsRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	chcl := res.Msg.Plants[0]
	if chcl.CompanyName != "Chilime Hydropower Company Limited" || *chcl.Close != 500 || *chcl.Pe != 20 ||
		chcl.Rating != "B" || chcl.Period != "Q1 2083/84" || chcl.CloseDate != "2026-11-04" {
		t.Fatalf("company figures not joined: %+v", chcl)
	}
	if sjcl := res.Msg.Plants[1]; sjcl.Close != nil || sjcl.Pe != nil || sjcl.Rating != "" {
		t.Fatalf("a plant without a company row must carry no figures: %+v", sjcl)
	}

	var basins []string
	for _, b := range res.Msg.Basins {
		basins = append(basins, fmt.Sprintf("%s=%v/%d", b.Name, b.State, b.PlantCount))
	}
	wantBasins := "Koshi=PLANT_STATE_UNSPECIFIED/0 Gandaki=PLANT_STATE_HIGH_DISCHARGE/2 Karnali=PLANT_STATE_UNSPECIFIED/0 " +
		"Mahakali=PLANT_STATE_UNSPECIFIED/0 Bagmati=PLANT_STATE_UNSPECIFIED/0 Kamala=PLANT_STATE_UNSPECIFIED/0 " +
		"West Rapti=PLANT_STATE_UNSPECIFIED/0 Babai=PLANT_STATE_UNSPECIFIED/0 Kankai=PLANT_STATE_NORMAL/1"
	if strings.Join(basins, " ") != wantBasins {
		t.Fatalf("basins %s", strings.Join(basins, " "))
	}
}

func TestGetCompanyPlantsFiltersBySymbol(t *testing.T) {
	svc := fixture(t)
	res, err := svc.GetCompanyPlants(
		context.Background(),
		connect.NewRequest(&ntxv1.GetCompanyPlantsRequest{Symbol: "SJCL"}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Msg.Plants) != 1 || res.Msg.Plants[0].Plant != "Sanjen" ||
		res.Msg.Plants[0].State != ntxv1.PlantState_PLANT_STATE_HIGH_DISCHARGE {
		t.Fatalf("got %+v", res.Msg.Plants)
	}
	if res.Msg.Opening != "One plant, 42.5 MW, on the Sanjen (Gandaki)." {
		t.Fatalf("opening %q", res.Msg.Opening)
	}
	none, err := svc.GetCompanyPlants(
		context.Background(),
		connect.NewRequest(&ntxv1.GetCompanyPlantsRequest{Symbol: "NABIL"}),
	)
	if err != nil || len(none.Msg.Plants) != 0 {
		t.Fatalf("a company without plants gets an empty list, got %v %v", none.Msg.Plants, err)
	}
}

func TestGetPlantReadingsIsNewestFirstAndSevenDeep(t *testing.T) {
	svc := fixture(t)
	res, err := svc.GetPlantReadings(
		context.Background(),
		connect.NewRequest(&ntxv1.GetPlantReadingsRequest{PlantId: 2}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Msg.Readings) != 7 || res.Msg.Readings[0].Day != day(0) || res.Msg.Readings[6].Day != day(-6) {
		t.Fatalf("got %d readings, first %s", len(res.Msg.Readings), res.Msg.Readings[0].Day)
	}
	r := res.Msg.Readings[0]
	if *r.DischargeM3S != 55 || *r.RainMm != 0 || r.RainSource != RainSource || r.DischargeSource != DischargeSource {
		t.Fatalf("today's reading %+v", r)
	}
	// Each day carries its own verdict: today's spike is high, yesterday's 20 m3/s is at the percentile, not above.
	if r.State != ntxv1.PlantState_PLANT_STATE_HIGH_DISCHARGE ||
		res.Msg.Readings[1].State != ntxv1.PlantState_PLANT_STATE_NORMAL {
		t.Fatalf("states %v %v", r.State, res.Msg.Readings[1].State)
	}
}
