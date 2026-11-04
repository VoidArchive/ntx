package worker

import "strings"

// tradedOn reports whether the market's last state change (asOf, RFC3339-ish
// local time) fell on businessDate (YYYY-MM-DD). Holidays leave asOf on the
// previous session, so the live feed's numbers must not be stored under today.
func tradedOn(asOf, businessDate string) bool {
	return businessDate != "" && strings.HasPrefix(asOf, businessDate)
}
