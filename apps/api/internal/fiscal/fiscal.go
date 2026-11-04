// Package fiscal is the single place NTX turns Gregorian dates into Nepali
// fiscal periods. It has no dependencies and no database.
//
// A fiscal year runs Shrawan 1 to Ashadh end (mid-July to mid-July) and is
// named by the Bikram Sambat year it starts in: Year(2082) prints as "2082/83".
// Quarters are Q1 Shrawan-Ashwin (the monsoon quarter), Q2 Kartik-Poush,
// Q3 Magh-Chaitra and Q4 Baisakh-Ashadh; Annual prints as "FY".
//
//	fy, err := fiscal.YearOf(date)       // Year, e.g. 2082/83
//	q, err := fiscal.QuarterOf(date)     // Q1..Q4
//	start, err := fy.Start()             // Shrawan 1 as a UTC midnight time.Time
//	end, err := fy.End()                 // Ashadh end
//	bs, err := fiscal.BSDate(date)       // "Kartik 18, 2083"
//	fy, err := fiscal.ParseYear(s)       // "2025-2026", "2080-2081", "2082/83" -> 2082/83, 2080/81, 2082/83
//	fy, err := fiscal.ParseFilingYear(np, gregorian) // Nepali name first, Gregorian span as fallback
//	q, err := fiscal.ParseQuarter(s)     // "First Quarter", "Q1" -> Q1; "Annual", "" -> Annual
//
// Dates are supported from Shrawan 1, 2075 (17 Jul 2018) up to but not
// including Shrawan 1, 2090 (16 Jul 2033); outside that the date functions
// return an error wrapping ErrOutOfTable. ParseYear and ParseQuarter only
// normalize names and accept any year.
package fiscal

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ErrOutOfTable is returned for dates and years the Shrawan 1 table does not cover.
var ErrOutOfTable = errors.New("fiscal: outside the Shrawan 1 table (2075-2090)")

// Year is a fiscal year, named by the Bikram Sambat year it starts in.
type Year int

// String prints the canonical form, e.g. "2082/83".
func (y Year) String() string {
	return fmt.Sprintf("%d/%02d", int(y), (int(y)+1)%100)
}

// Start is Shrawan 1 of the fiscal year, at UTC midnight.
func (y Year) Start() (time.Time, error) {
	i := int(y) - firstYear
	if i < 0 || i >= len(years)-1 {
		return time.Time{}, fmt.Errorf("%w: fiscal year %s", ErrOutOfTable, y)
	}
	return years[i].shrawan1, nil
}

// End is the last calendar day of the fiscal year (Ashadh end), at UTC midnight.
func (y Year) End() (time.Time, error) {
	i := int(y) - firstYear + 1
	if i < 1 || i >= len(years) {
		return time.Time{}, fmt.Errorf("%w: fiscal year %s", ErrOutOfTable, y)
	}
	return years[i].shrawan1.AddDate(0, 0, -1), nil
}

// Quarter is a fiscal quarter, or Annual for a full-year report.
type Quarter int

const (
	Annual Quarter = iota
	Q1
	Q2
	Q3
	Q4
)

// String prints the canonical form: "Q1".."Q4", or "FY" for Annual.
func (q Quarter) String() string {
	if q == Annual {
		return "FY"
	}
	return fmt.Sprintf("Q%d", int(q))
}

// YearOf returns the fiscal year containing d. Only the calendar day of d is
// used, whatever its zone.
func YearOf(d time.Time) (Year, error) {
	i, _, err := locate(d)
	if err != nil {
		return 0, err
	}
	return Year(firstYear + i), nil
}

// QuarterOf returns the fiscal quarter containing d.
func QuarterOf(d time.Time) (Quarter, error) {
	i, offset, err := locate(d)
	if err != nil {
		return 0, err
	}
	// Quarters end at Ashwin, Poush and Chaitra end: months 6, 9 and 12 of the start year.
	q, end := Q1, 0
	for m := 3; m < 12; m++ {
		end += years[i].months[m]
		if m%3 == 2 && offset >= end {
			q++
		}
	}
	return q, nil
}

// BSDate prints d as a Bikram Sambat date, e.g. "Kartik 18, 2083".
func BSDate(d time.Time) (string, error) {
	i, offset, err := locate(d)
	if err != nil {
		return "", err
	}
	// Walk Shrawan..Chaitra of the start year, then Baisakh..Ashadh of the next.
	for k := range 12 {
		m, yr := (3+k)%12, i
		if k >= 9 {
			yr++
		}
		if offset < years[yr].months[m] {
			return fmt.Sprintf("%s %d, %d", monthNames[m], offset+1, firstYear+yr), nil
		}
		offset -= years[yr].months[m]
	}
	return "", fmt.Errorf("%w: %s", ErrOutOfTable, d.Format(time.DateOnly))
}

// locate returns the table index of the fiscal year containing d and d's day
// offset from that year's Shrawan 1.
func locate(d time.Time) (int, int, error) {
	y, m, dd := d.Date()
	day := time.Date(y, m, dd, 0, 0, 0, 0, time.UTC)
	last := len(years) - 1
	if day.Before(years[0].shrawan1) || !day.Before(years[last].shrawan1) {
		return 0, 0, fmt.Errorf("%w: %s", ErrOutOfTable, day.Format(time.DateOnly))
	}
	i := last - 1
	for day.Before(years[i].shrawan1) {
		i--
	}
	return i, int(day.Sub(years[i].shrawan1).Hours() / 24), nil
}

var yearSpan = regexp.MustCompile(`^(\d{4})[/-](\d{2}|\d{4})$`)

// ParseYear normalizes a fiscal-year name to its Year. It accepts the canonical
// "2082/83", the exchange's hyphenated BS form "2082-2083" and its Gregorian
// span "2025-2026" (the fiscal year starting in Gregorian 2025 is BS 2082/83).
func ParseYear(s string) (Year, error) {
	m := yearSpan.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, fmt.Errorf("fiscal: %q is not a fiscal year", s)
	}
	first, _ := strconv.Atoi(m[1])
	second, _ := strconv.Atoi(m[2])
	if second%100 != (first+1)%100 {
		return 0, fmt.Errorf("fiscal: %q does not span consecutive years", s)
	}
	// BS 2057 began in Gregorian 2000, so any four-digit year below that is Gregorian.
	if first < 2057 {
		return Year(first + 57), nil
	}
	return Year(first), nil
}

// ParseFilingYear reads a filing's year from the exchange's Nepali name, and
// from its Gregorian span when the Nepali name is missing or unreadable.
func ParseFilingYear(nepali, gregorian string) (Year, error) {
	if y, err := ParseYear(nepali); err == nil {
		return y, nil
	}
	return ParseYear(gregorian)
}

// ParseQuarter normalizes a quarter name to its Quarter. It accepts "Q1".."Q4",
// the exchange's "First Quarter".."Fourth Quarter", and "Annual" or the empty
// string the exchange uses for annual reports.
func ParseQuarter(s string) (Quarter, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "annual", "annual report", "fy":
		return Annual, nil
	case "q1", "first quarter":
		return Q1, nil
	case "q2", "second quarter":
		return Q2, nil
	case "q3", "third quarter":
		return Q3, nil
	case "q4", "fourth quarter":
		return Q4, nil
	}
	return 0, fmt.Errorf("fiscal: %q is not a quarter", s)
}
