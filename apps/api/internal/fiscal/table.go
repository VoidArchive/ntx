package fiscal

import "time"

// firstYear is the BS year of the first table row. The table ends with the
// Shrawan 1 of BS 2090, which is the exclusive upper bound of supported dates.
const firstYear = 2075

// bsYear holds one Bikram Sambat year: the Gregorian date of its Shrawan 1
// (fiscal-year start) and the lengths of its twelve months, Baisakh first.
type bsYear struct {
	shrawan1 time.Time
	months   [12]int
}

// july is the Gregorian date of a Shrawan 1; it always falls on 16 or 17 July.
func july(y, d int) time.Time {
	return time.Date(y, time.July, d, 0, 0, 0, 0, time.UTC)
}

// years is indexed by BS year minus firstYear.
//
// Sources, checked by hand on 2026-11-04:
//   - Hamro Patro month pages, https://www.hamropatro.com/calendar/<bs-year>/<month>
//     (e.g. https://www.hamropatro.com/calendar/2082/4 shows Shrawan 1, 2082 = 17 Jul 2025).
//     Every month length and first-of-month date for 2075-2086 below matches those pages.
//   - nepali-datetime's table, https://raw.githubusercontent.com/arneec/nepali-datetime/master/nepali_datetime/data/calendar_bs.csv,
//     agrees with Hamro Patro for 2075-2086 and is the source of the 2087-2090 rows.
//
// BS 2087-2090 are projections: the Panchanga committee has not published them,
// and Hamro Patro's own 2087 page sums to 367 days (Mangsir 30 instead of 29),
// which pushes its Shrawan 1, 2088 to 18 Jul. The nepali-datetime rows keep every
// Shrawan 1 on 16 or 17 Jul and are used instead. Re-check those four rows when the
// official calendars appear.
var years = [...]bsYear{
	{july(2018, 17), [12]int{31, 31, 32, 31, 31, 31, 30, 29, 30, 29, 30, 30}}, // 2075
	{july(2019, 17), [12]int{31, 32, 31, 32, 31, 30, 30, 30, 29, 29, 30, 30}}, // 2076
	{july(2020, 16), [12]int{31, 32, 31, 32, 31, 30, 30, 30, 29, 30, 29, 31}}, // 2077
	{july(2021, 16), [12]int{31, 31, 31, 32, 31, 31, 30, 29, 30, 29, 30, 30}}, // 2078
	{july(2022, 17), [12]int{31, 31, 32, 31, 31, 31, 30, 29, 30, 29, 30, 30}}, // 2079
	{july(2023, 17), [12]int{31, 32, 31, 32, 31, 30, 30, 30, 29, 29, 30, 30}}, // 2080
	{july(2024, 16), [12]int{31, 32, 31, 32, 31, 30, 30, 30, 29, 30, 29, 31}}, // 2081
	{july(2025, 17), [12]int{31, 31, 32, 31, 31, 31, 30, 29, 30, 29, 30, 30}}, // 2082
	{july(2026, 17), [12]int{31, 31, 32, 31, 31, 31, 30, 29, 30, 29, 30, 30}}, // 2083
	{july(2027, 17), [12]int{31, 31, 32, 31, 31, 30, 30, 30, 29, 30, 30, 30}}, // 2084
	{july(2028, 16), [12]int{31, 32, 31, 32, 30, 31, 30, 30, 29, 30, 30, 30}}, // 2085
	{july(2029, 16), [12]int{30, 32, 31, 32, 31, 30, 30, 30, 29, 30, 30, 30}}, // 2086
	{july(2030, 17), [12]int{31, 31, 32, 31, 31, 31, 30, 29, 30, 30, 30, 30}}, // 2087 (projected)
	{july(2031, 17), [12]int{30, 31, 32, 32, 30, 31, 30, 30, 29, 30, 30, 30}}, // 2088 (projected)
	{july(2032, 16), [12]int{30, 32, 31, 32, 31, 30, 30, 30, 29, 30, 30, 30}}, // 2089 (projected)
	{july(2033, 16), [12]int{30, 32, 31, 32, 31, 30, 30, 30, 29, 30, 30, 30}}, // 2090 (projected)
}

var monthNames = [12]string{
	"Baisakh", "Jestha", "Ashadh", "Shrawan", "Bhadra", "Ashwin",
	"Kartik", "Mangsir", "Poush", "Magh", "Falgun", "Chaitra",
}
