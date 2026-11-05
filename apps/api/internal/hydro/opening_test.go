package hydro

import (
	"testing"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
)

func TestCompanyOpening(t *testing.T) {
	plant := func(river, basin string, mw float64) *ntxv1.Plant {
		return &ntxv1.Plant{River: river, Basin: basin, Mw: mw}
	}
	cases := []struct {
		name   string
		plants []*ntxv1.Plant
		want   string
	}{
		{"none", nil, ""},
		{"one", []*ntxv1.Plant{plant("Chilime", "Gandaki", 22)}, "One plant, 22 MW, on the Chilime (Gandaki)."},
		{
			"two on one river",
			[]*ntxv1.Plant{plant("Trishuli", "Gandaki", 24), plant("Trishuli", "Gandaki", 12)},
			"Two plants, 36 MW, on the Trishuli (Gandaki).",
		},
		{
			"two rivers one basin",
			[]*ntxv1.Plant{plant("Sanjen", "Gandaki", 42.5), plant("Chilime", "Gandaki", 22)},
			"Two plants, 64.5 MW, on the Sanjen and the Chilime (Gandaki).",
		},
		{
			"two basins",
			[]*ntxv1.Plant{plant("Iwa Khola", "Koshi", 9.9), plant("Ridi", "Gandaki", 2.4)},
			"Two plants, 12.3 MW, on the Iwa Khola (Koshi) and the Ridi (Gandaki).",
		},
		{"many", []*ntxv1.Plant{
			plant("Naugad", "Mahakali", 8.5), plant("Chameliya Khola", "Mahakali", 40), plant("Naugad", "Mahakali", 8),
		}, "Three plants, 56.5 MW, on the Naugad and the Chameliya Khola (Mahakali)."},
	}
	for _, c := range cases {
		if got := CompanyOpening(c.plants); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestMapOpening(t *testing.T) {
	plants := []*ntxv1.Plant{
		{Symbol: "CHCL", River: "Chilime", Basin: "Gandaki", Latest: &ntxv1.Reading{Day: "2026-11-05"}},
		{Symbol: "SJCL", River: "Sanjen", Basin: "Gandaki", Latest: &ntxv1.Reading{Day: "2026-11-05"}},
		{Symbol: "SHPC", River: "Mai Khola", Basin: "Kankai", Latest: &ntxv1.Reading{Day: "2026-10-31"}},
		{Symbol: "SHPC", River: "Mai Khola", Basin: "Kankai"},
	}
	basins := []*ntxv1.Basin{
		{Name: "Koshi"},
		{Name: "Gandaki", State: ntxv1.PlantState_PLANT_STATE_HIGH_DISCHARGE, PlantCount: 2},
		{Name: "Kankai", State: ntxv1.PlantState_PLANT_STATE_NORMAL, PlantCount: 2},
	}
	want := "4 plants of 3 listed companies on 3 rivers. Basins on 2026-11-05: Gandaki high discharge, Kankai normal."
	if got := MapOpening(plants, basins); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
	if got := MapOpening(plants[3:], basins); got != "1 plants of 1 listed companies on 1 rivers. No readings yet." {
		t.Errorf("without readings: %q", got)
	}
}
