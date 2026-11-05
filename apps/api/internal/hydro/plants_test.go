package hydro

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/voidarchive/ntx/internal/database"
)

const header = "symbol,plant,river,basin,district,mw,lat,lon,status,source_url\n"

const goodRow = "CHCL,Chilime,Chilime,Gandaki,Rasuwa,22,28.1698,85.3196,operating,https://www.doed.gov.np/pages/powerplantsmorethan1\n"

func TestReadPlantsRejectsBadRows(t *testing.T) {
	cases := map[string]string{
		"unknown basin":          "CHCL,Chilime,Chilime,Trishuli,Rasuwa,22,28.1,85.3,operating,https://doed.gov.np\n",
		"unknown status":         "CHCL,Chilime,Chilime,Gandaki,Rasuwa,22,28.1,85.3,planned,https://doed.gov.np\n",
		"source_url is required": "CHCL,Chilime,Chilime,Gandaki,Rasuwa,22,28.1,85.3,operating,\n",
		"outside Nepal":          "CHCL,Chilime,Chilime,Gandaki,Rasuwa,22,48.1,85.3,operating,https://doed.gov.np\n",
		"positive number":        "CHCL,Chilime,Chilime,Gandaki,Rasuwa,0,28.1,85.3,operating,https://doed.gov.np\n",
		"header must be":         "symbol,plant\nCHCL,Chilime\n",
	}
	for want, csv := range cases {
		in := csv
		if !strings.HasPrefix(in, "symbol") {
			in = header + goodRow + csv
		}
		_, err := ReadPlants(strings.NewReader(in))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got err %v", want, err)
		}
		if err != nil && !strings.HasPrefix(in, "symbol,plant\n") && !strings.Contains(err.Error(), "row 3") {
			t.Errorf("%s: error should name row 3, got %v", want, err)
		}
	}
}

func TestReadPlantsParsesARow(t *testing.T) {
	plants, err := ReadPlants(strings.NewReader(header + goodRow))
	if err != nil {
		t.Fatal(err)
	}
	p := plants[0]
	if p.Symbol != "CHCL" || p.Basin != "Gandaki" || p.Mw != 22 || p.Lat != 28.1698 || p.Lon != 85.3196 {
		t.Fatalf("parsed %+v", p)
	}
}

func TestSyncPlantsUpsertsBySymbolAndPlant(t *testing.T) {
	ctx := context.Background()
	q := database.OpenTestDB(t)
	path := filepath.Join(t.TempDir(), "plants.csv")
	write := func(mw string) {
		t.Helper()
		row := strings.Replace(goodRow, ",22,", ","+mw+",", 1)
		if err := os.WriteFile(path, []byte(header+row), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := SyncPlants(ctx, q, path); err != nil {
			t.Fatal(err)
		}
	}
	write("22")
	write("22.1")
	plants, err := q.ListHydroPlantsBySymbol(ctx, "CHCL")
	if err != nil {
		t.Fatal(err)
	}
	if len(plants) != 1 || plants[0].Mw != 22.1 {
		t.Fatalf("got %+v, want one row at 22.1 MW", plants)
	}
	if err := SyncPlants(ctx, q, filepath.Join(t.TempDir(), "missing.csv")); err != nil {
		t.Fatalf("a missing file should be skipped, got %v", err)
	}
}

// The committed table must load, so a hand edit that breaks a row fails CI.
func TestRepoPlantsLoad(t *testing.T) {
	f, err := os.Open("../../../../data/hydro_plants.csv")
	if err != nil {
		t.Skip("data/hydro_plants.csv not present")
	}
	defer f.Close()
	plants, err := ReadPlants(f)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, p := range plants {
		key := p.Symbol + "\x00" + p.Plant
		if seen[key] {
			t.Errorf("duplicate plant %s %q", p.Symbol, p.Plant)
		}
		seen[key] = true
	}
}
