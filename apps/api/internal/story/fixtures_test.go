package story

import (
	"database/sql"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// Rows captured from ~/.local/share/ntx/market.db on 2026-11-04.

type fixtureRow struct {
	fy, quarter                    string
	eps, pe, bv, paidUp, profitYTD float64
}

var nabilRows = []fixtureRow{
	{"2019-2020", "", 37.24, 20.54, 255.45, 13480158400.0, 3570000000.0},
	{"2021-2022", "Fourth Quarter", 26.88, 32.8, 230.37, 22832908632.34, 4972756000.0},
	{"2021-2022", "", 18.64, 44.21, 232.0, 22832908632.0, 4256024188.0},
	{"2022-2023", "First Quarter", 27.01, 29.39, 233.28, 22832908632.0, 1571537000.0},
	{"2022-2023", "Second Quarter", 25.33, 28.19, 199.12, 27056997000.0, 3417698000.0},
	{"2022-2023", "Third Quarter", 25.21, 24.55, 201.07, 27056997000.0, 5116174000.0},
	{"2022-2023", "Fourth Quarter", 27.82, 21.54, 214.69, 27056997000.0, 7527149000.0},
	{"2022-2023", "", 23.67, 25.31, 210.0, 27056996729.0, 6404936160.0},
	{"2023-2024", "First Quarter", 21.31, 26.84, 218.27, 27056997000.0, 14693000.0},
	{"2023-2024", "Second Quarter", 23.74, 20.83, 204.44, 27056996729.0, 3203079000.0},
	{"2023-2024", "Third Quarter", 23.24, 18.85, 208.39, 27056996729.0, 4668029000.0},
	{"2023-2024", "Fourth Quarter", 26.09, 20.08, 220.16, 27056996729.0, 7060033000.0},
	{"2023-2024", "", 22.9, 22.88, 216.86, 27056996729.0, 6197108561.0},
	{"2024-2025", "First Quarter", 29.83, 20.38, 229.48, 27056996729.0, 2056590000.0},
	{"2024-2025", "Second Quarter", 24.05, 20.36, 219.01, 27056996729.0, 3244277000.0},
	{"2024-2025", "Third Quarter", 25.05, 19.44, 226.36, 27056996729.0, 5051419000.0},
	{"2024-2025", "Fourth Quarter", 26.34, 20.55, 235.58, 27056996729.0, 7127566000.0},
	{"2024-2025", "", 21.89, 24.73, 231.0, 27056996729.0, 5922144047.0},
	{"2025-2026", "First Quarter", 25.49, 19.97, 240.72, 27056996729.0, 1756991000.0},
	{"2025-2026", "Second Quarter", 35.04, 13.91, 235.64, 27056996729.0, 4759641000.0},
	{"2025-2026", "Third Quarter", 33.02, 16.05, 243.3, 27056996729.0, 6764646000.0},
	{"2025-2026", "Fourth Quarter", 28.36, 18.86, 247.28, 27056996729.0, 7905796000.0},
	{"2026-2027", "First Quarter", 33.89, 16.14, 233.9, 27056996729.0, 2555105000.0},
}

var apiRows = []fixtureRow{
	{"2018-2019", "", 5.96, 24.01, 106.23, 1190700000.0, 70922011.0},
	{"2019-2020", "", 8.04, 13.43, 114.28, 2496718400.0, 95779743.0},
	{"2020-2021", "", 5.09, 79.12, 111.31, 2475196250.0, 126074668.0},
	{"2021-2022", "First Quarter", 11.12, 38.5, 106.01, 2758915546.0, 76679590.0},
	{"2021-2022", "Second Quarter", 10.37, 50.72, 108.42, 2758915546.0, 143050116.0},
	{"2021-2022", "Third Quarter", 8.4, 36.95, 109.54, 2758915546.0, 173891301.0},
	{"2021-2022", "Fourth Quarter", 3.67, 66.66, 107.72, 3844915546.0, 141018788.0},
	{"2021-2022", "", 3.06, 79.0, 108.0, 3844915546.0, 106142234.0},
	{"2022-2023", "First Quarter", 4.55, 45.6, 108.85, 3844915546.0, 43744330.0},
	{"2022-2023", "Second Quarter", 3.25, 84.12, 109.66, 3844915546.0, 62393513.0},
	{"2022-2023", "Third Quarter", 1.62, 158.64, 101.2, 4133284212.0, 50302924.0},
	{"2022-2023", "Fourth Quarter", 3.94, 50.01, 103.93, 4133284209.0, 162809640.0},
	{"2022-2023", "", 1.07, 184.11, 101.28, 4133284212.0, 44238624.0},
	{"2023-2024", "First Quarter", -0.29, -593.79, 102.45, 5667104512.0, -4049241.0},
	{"2023-2024", "Second Quarter", 2.75, 70.65, 103.31, 5786597912.0, 79453420.0},
	{"2023-2024", "Third Quarter", 4.25, 41.44, 105.2, 5786579912.0, 184303655.0},
	{"2023-2024", "Fourth Quarter", 7.89, 24.21, 109.73, 5786597912.0, 456584163.0},
	{"2024-2025", "First Quarter", 12.56, 24.83, 112.65, 5786597912.0, 181703396.0},
	{"2024-2025", "Second Quarter", 10.26, 25.83, 109.18, 6075927807.0, 311688350.0},
	{"2024-2025", "Third Quarter", 7.76, 35.2, 109.85, 6075927807.0, 353497566.0},
	{"2024-2025", "Fourth Quarter", 7.04, 42.04, 111.07, 6075927807.0, 427659228.0},
	{"2025-2026", "First Quarter", 17.3, 15.62, 114.74, 6075927807.0, 262755150.0},
	{"2025-2026", "Second Quarter", 14.03, 20.62, 117.58, 6075927807.0, 426221665.0},
	{"2025-2026", "Third Quarter", 13.76, 25.3, 115.32, 6379724199.0, 658546633.0},
	{"2025-2026", "Fourth Quarter", 15.06, 21.3, 119.51, 6379724199.0, 960605093.0},
}

// Commercial Banks at their latest quarterly row and the 2026-11-03 close,
// as close/eps and close/book_value (NIMB's book value is bad upstream data),
// plus API's 2080/81 Q1 loss to stand in for a loss-maker.
var bankSector = Sector{
	PE: []float64{
		111.4754, 83.5821, 20.414, 11.6488, 31.0976, 14.5027, 23.7887, 15.4658, 16.288, 13.638,
		105.0746, 15.3791, 11.7079, 14.1111, 18.5185, 15.8447, 22.2011, 19.1631, 25.1311, -593.79,
	},
	PB: []float64{
		1.461863, 1.441163, 2.92169, 1.547477, 1.149425, 1.37949, 1.56423, 1.666058, 2.359983, 1.054662,
		2.04058, 0.0000000031, 1.497424, 1.575654, 1.322751, 2.165375, 2.167184, 1.962517, 2.890995,
	},
	EPS: []float64{
		3.05, 2.68, 34.78, 23.35, 6.56, 16.39, 11.64, 17.71, 33.89, 22.65,
		3.35, 13.85, 23.83, 19.63, 10.26, 25.75, 19.49, 23.06, 26.7,
	},
}

func fundamentals(companyID int64, rows []fixtureRow) []sqlc.Fundamental {
	out := make([]sqlc.Fundamental, 0, len(rows))
	for _, r := range rows {
		out = append(out, sqlc.Fundamental{
			CompanyID:     companyID,
			FiscalYear:    r.fy,
			Quarter:       r.quarter,
			Eps:           num(r.eps),
			PeRatio:       num(r.pe),
			BookValue:     num(r.bv),
			PaidUpCapital: num(r.paidUp),
			ProfitAmount:  num(r.profitYTD),
		})
	}
	return out
}

// asOf keeps the rows a reader would have seen right after the given quarter was filed.
func asOf(rows []fixtureRow, fy, quarter string) []fixtureRow {
	rank := map[string]int{"First Quarter": 1, "Second Quarter": 2, "Third Quarter": 3, "Fourth Quarter": 4, "": 5}
	var out []fixtureRow
	for _, r := range rows {
		if r.fy < fy || (r.fy == fy && rank[r.quarter] <= rank[quarter]) {
			out = append(out, r)
		}
	}
	return out
}

func num(v float64) sql.NullFloat64 { return sql.NullFloat64{Float64: v, Valid: true} }

func price(date string, close float64) sqlc.Price {
	return sqlc.Price{BusinessDate: date, ClosePrice: num(close)}
}

var (
	nabil = sqlc.Company{ID: 131, Name: "Nabil Bank Limited", Symbol: "NABIL", Sector: "Commercial Banks"}
	api   = sqlc.Company{ID: 697, Name: "API POWER COMPANY LIMITED", Symbol: "API", Sector: "Hydro Power"}
)
