package nepse

// SectorMap maps NEPSE sector names to database integer IDs.
var SectorMap = map[string]int64{
	"Commercial Banks":             1,
	"Development Banks":            2,
	"Finance":                      3,
	"Microfinance":                 4,
	"Life Insurance":               5,
	"Non Life Insurance":           6,
	"Hydro Power":                  7,
	"Manufacturing And Processing": 8,
	"Hotels And Tourism":           9,
	"Trading":                      10,
	"Investment":                   11,
	"Mutual Fund":                  12,
	"Others":                       13,
}

// SectorToInt converts a sector name to its database integer ID.
// Returns 0 if the sector name is not recognized.
func SectorToInt(name string) int64 {
	if id, ok := SectorMap[name]; ok {
		return id
	}
	return 0
}
