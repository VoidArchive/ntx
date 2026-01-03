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

// SectorNames maps database integer IDs to sector names.
var SectorNames = map[int64]string{
	1:  "Commercial Banks",
	2:  "Development Banks",
	3:  "Finance",
	4:  "Microfinance",
	5:  "Life Insurance",
	6:  "Non Life Insurance",
	7:  "Hydro Power",
	8:  "Manufacturing And Processing",
	9:  "Hotels And Tourism",
	10: "Trading",
	11: "Investment",
	12: "Mutual Fund",
	13: "Others",
}

// SectorToInt converts a sector name to its database integer ID.
// Returns 13 (Others) if the sector name is not recognized.
func SectorToInt(name string) int64 {
	if id, ok := SectorMap[name]; ok {
		return id
	}
	return 13 // Others
}

// SectorToName converts a sector ID to its name.
// Returns "Unknown" if the ID is not recognized.
func SectorToName(id int64) string {
	if name, ok := SectorNames[id]; ok {
		return name
	}
	return "Unknown"
}
