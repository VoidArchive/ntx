package hydro

import (
	"fmt"
	"strconv"
	"strings"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
)

// stateWords are the prose forms of the states; the enum names are for code.
var stateWords = map[ntxv1.PlantState]string{
	ntxv1.PlantState_PLANT_STATE_UNSPECIFIED:    "no readings",
	ntxv1.PlantState_PLANT_STATE_NORMAL:         "normal",
	ntxv1.PlantState_PLANT_STATE_HEAVY_RAIN:     "heavy rain",
	ntxv1.PlantState_PLANT_STATE_HIGH_DISCHARGE: "high discharge",
}

// MapOpening is the /hydro page's first sentence: counts, then each basin with
// plants and its state on the newest reading day.
func MapOpening(plants []*ntxv1.Plant, basins []*ntxv1.Basin) string {
	companies, rivers := map[string]bool{}, map[string]bool{}
	day := ""
	for _, p := range plants {
		companies[p.Symbol] = true
		rivers[p.River] = true
		day = max(day, p.Latest.GetDay())
	}
	s := fmt.Sprintf("%d plants of %d listed companies on %d rivers.", len(plants), len(companies), len(rivers))
	if day == "" {
		return s + " No readings yet."
	}
	var states []string
	for _, b := range basins {
		if b.PlantCount > 0 {
			states = append(states, b.Name+" "+stateWords[b.State])
		}
	}
	return s + " Basins on " + day + ": " + strings.Join(states, ", ") + "."
}

// CompanyOpening heads the company page's Plants section:
// "Two plants, 36 MW, on the Trishuli (Gandaki)."
func CompanyOpening(plants []*ntxv1.Plant) string {
	if len(plants) == 0 {
		return ""
	}
	mw := 0.0
	var rivers []string
	basins := map[string]bool{}
	seen := map[string]bool{}
	for _, p := range plants {
		mw += p.Mw
		basins[p.Basin] = true
		if !seen[p.River] {
			seen[p.River] = true
			rivers = append(rivers, p.River)
		}
	}
	// One basin is named once at the end; several are named beside each river.
	names := make([]string, len(rivers))
	for i, r := range rivers {
		names[i] = "the " + r
		if len(basins) > 1 {
			names[i] += " (" + basinOf(plants, r) + ")"
		}
	}
	where := strings.Join(names[:len(names)-1], ", ")
	if where != "" {
		where += " and "
	}
	where += names[len(names)-1]
	if len(basins) == 1 {
		where += " (" + plants[0].Basin + ")"
	}
	return fmt.Sprintf(
		"%s, %s MW, on %s.",
		countWord(len(plants), "plant"),
		strconv.FormatFloat(mw, 'f', -1, 64),
		where,
	)
}

func basinOf(plants []*ntxv1.Plant, river string) string {
	for _, p := range plants {
		if p.River == river {
			return p.Basin
		}
	}
	return ""
}

var smallNumbers = []string{"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine"}

func countWord(n int, noun string) string {
	word := strconv.Itoa(n)
	if n < len(smallNumbers) {
		word = smallNumbers[n]
	}
	if n == 1 {
		return word + " " + noun
	}
	return word + " " + noun + "s"
}
