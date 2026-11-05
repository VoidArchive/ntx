// Package hydro holds the plant table, the daily weather readings and the two
// state rules behind /hydro (GLOSSARY: Plant, Reading, Plant state; ADR 0004).
package hydro

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// Basins are the nine drainage basins of GLOSSARY.md, in display order.
var Basins = []string{"Koshi", "Gandaki", "Karnali", "Mahakali", "Bagmati", "Kamala", "West Rapti", "Babai", "Kankai"}

var statuses = []string{"operating", "under_construction"}

var columns = []string{"symbol", "plant", "river", "basin", "district", "mw", "lat", "lon", "status", "source_url"}

type Plant = sqlc.UpsertHydroPlantParams

// ReadPlants parses data/hydro_plants.csv. A row with an unknown basin or status,
// no source URL, or a point outside Nepal fails the whole load: a bad row must
// not reach the map.
func ReadPlants(r io.Reader) ([]Plant, error) {
	rows, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 || !slices.Equal(rows[0], columns) {
		return nil, fmt.Errorf("header must be %s", strings.Join(columns, ","))
	}
	plants := make([]Plant, 0, len(rows)-1)
	for i, rec := range rows[1:] {
		p, err := parsePlant(rec)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+2, err)
		}
		plants = append(plants, p)
	}
	return plants, nil
}

func parsePlant(rec []string) (Plant, error) {
	p := Plant{
		Symbol: rec[0], Plant: rec[1], River: rec[2], Basin: rec[3], District: rec[4],
		Status: rec[8], SourceUrl: rec[9],
	}
	if p.Symbol == "" || p.Plant == "" || p.River == "" {
		return p, errors.New("symbol, plant and river are required")
	}
	if !slices.Contains(Basins, p.Basin) {
		return p, fmt.Errorf("unknown basin %q", p.Basin)
	}
	if !slices.Contains(statuses, p.Status) {
		return p, fmt.Errorf("unknown status %q", p.Status)
	}
	if !strings.HasPrefix(p.SourceUrl, "http") {
		return p, errors.New("source_url is required (ADR 0002)")
	}
	var err error
	if p.Mw, err = strconv.ParseFloat(rec[5], 64); err != nil || p.Mw <= 0 {
		return p, fmt.Errorf("mw %q must be a positive number", rec[5])
	}
	if p.Lat, err = strconv.ParseFloat(rec[6], 64); err != nil || p.Lat < 26 || p.Lat > 31 {
		return p, fmt.Errorf("lat %q is outside Nepal", rec[6])
	}
	if p.Lon, err = strconv.ParseFloat(rec[7], 64); err != nil || p.Lon < 80 || p.Lon > 89 {
		return p, fmt.Errorf("lon %q is outside Nepal", rec[7])
	}
	return p, nil
}

// SyncPlants upserts the CSV at path, keyed on (symbol, plant). A missing file is
// logged and skipped so a checkout without the data still serves every other page.
// shortcut: a row removed from the CSV keeps its table row and readings; delete it by hand.
func SyncPlants(ctx context.Context, q *sqlc.Queries, path string) error {
	f, err := os.Open(path) //nolint:gosec // the path is configuration, not user input
	if errors.Is(err, fs.ErrNotExist) {
		slog.Warn("hydro plants csv missing; /hydro will be empty", "path", path)
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	plants, err := ReadPlants(f)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	for _, p := range plants {
		if err := q.UpsertHydroPlant(ctx, p); err != nil {
			return fmt.Errorf("upsert plant %s %q: %w", p.Symbol, p.Plant, err)
		}
	}
	slog.Info("hydro plants synced", "plants", len(plants))
	return nil
}
