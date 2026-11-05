package hydro

import "database/sql"

// nullFloat keeps a zero: 0 mm of rain is a reading, unlike the worker's price helpers.
func nullFloat(v float64) sql.NullFloat64 { return sql.NullFloat64{Float64: v, Valid: true} }

func nullString(s string) sql.NullString { return sql.NullString{String: s, Valid: s != ""} }
