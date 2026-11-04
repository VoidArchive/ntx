package fiscal

import (
	"errors"
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestYearOf(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want string
		err  bool
	}{
		{"day before 2082/83 starts", day(2025, time.July, 16), "2081/82", false},
		{"first day of 2082/83", day(2025, time.July, 17), "2082/83", false},
		{"last day of 2082/83", day(2026, time.July, 16), "2082/83", false},
		{"first day of 2083/84", day(2026, time.July, 17), "2083/84", false},
		{"first day of the table", day(2018, time.July, 17), "2075/76", false},
		{"last day of the table", day(2033, time.July, 15), "2089/90", false},
		{"before the table", day(2018, time.July, 16), "", true},
		{"past the table", day(2033, time.July, 16), "", true},
		// Callers pass business dates from any zone; only the calendar day matters.
		{
			"non-UTC time on a boundary",
			time.Date(2025, time.July, 17, 23, 30, 0, 0, time.FixedZone("NPT", 5*3600+45*60)),
			"2082/83",
			false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := YearOf(tc.date)
			if tc.err {
				if !errors.Is(err, ErrOutOfTable) {
					t.Fatalf("want ErrOutOfTable, got %v (%v)", err, got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.String() != tc.want {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestQuarterOf(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want Quarter
	}{
		{"monsoon quarter", day(2025, time.August, 15), Q1},
		{"Ashwin end", day(2025, time.October, 17), Q1},
		{"Kartik 1", day(2025, time.October, 18), Q2},
		{"Poush end", day(2026, time.January, 14), Q2},
		{"Magh 1", day(2026, time.January, 15), Q3},
		{"Chaitra end", day(2026, time.April, 13), Q3},
		{"Baisakh 1", day(2026, time.April, 14), Q4},
		{"Ashadh end", day(2026, time.July, 16), Q4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := QuarterOf(tc.date)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
	if _, err := QuarterOf(day(2033, time.July, 16)); !errors.Is(err, ErrOutOfTable) {
		t.Fatalf("want ErrOutOfTable, got %v", err)
	}
}

func TestYearBounds(t *testing.T) {
	start, err := Year(2082).Start()
	if err != nil || !start.Equal(day(2025, time.July, 17)) {
		t.Fatalf("Start: got %v, %v", start, err)
	}
	end, err := Year(2082).End()
	if err != nil || !end.Equal(day(2026, time.July, 16)) {
		t.Fatalf("End: got %v, %v", end, err)
	}
	end, err = Year(2089).End()
	if err != nil || !end.Equal(day(2033, time.July, 15)) {
		t.Fatalf("End 2089: got %v, %v", end, err)
	}
	if _, err := Year(2090).End(); !errors.Is(err, ErrOutOfTable) {
		t.Fatalf("End 2090: want ErrOutOfTable, got %v", err)
	}
	if _, err := Year(2074).Start(); !errors.Is(err, ErrOutOfTable) {
		t.Fatalf("Start 2074: want ErrOutOfTable, got %v", err)
	}
}

func TestBSDate(t *testing.T) {
	tests := []struct {
		date time.Time
		want string
	}{
		{day(2025, time.July, 17), "Shrawan 1, 2082"},
		{day(2025, time.October, 17), "Ashwin 31, 2082"},
		{day(2026, time.January, 15), "Magh 1, 2082"},
		{day(2026, time.April, 14), "Baisakh 1, 2083"},
		{day(2026, time.July, 16), "Ashadh 32, 2083"},
		{day(2026, time.November, 4), "Kartik 18, 2083"},
		{day(2033, time.July, 15), "Ashadh 31, 2090"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got, err := BSDate(tc.date)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
	if _, err := BSDate(day(2018, time.July, 16)); !errors.Is(err, ErrOutOfTable) {
		t.Fatalf("want ErrOutOfTable, got %v", err)
	}
}

func TestParseYear(t *testing.T) {
	tests := []struct {
		in   string
		want string
		err  bool
	}{
		{"2025-2026", "2082/83", false}, // exchange FYName, Gregorian span
		{"2080-2081", "2080/81", false}, // exchange FYNameNepali, hyphenated BS
		{"2082/83", "2082/83", false},   // already canonical
		{" 2082/2083 ", "2082/83", false},
		{"2025-2027", "", true},
		{"FY 2082", "", true},
		{"", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseYear(tc.in)
			if tc.err {
				if err == nil {
					t.Fatalf("want error, got %s", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.String() != tc.want {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestParseQuarter(t *testing.T) {
	tests := []struct {
		in   string
		want Quarter
		err  bool
	}{
		{"First Quarter", Q1, false},
		{"Second Quarter", Q2, false},
		{"Third Quarter", Q3, false},
		{"Fourth Quarter", Q4, false},
		{"q3", Q3, false},
		{"Annual", Annual, false},
		{"", Annual, false}, // the exchange stores annual reports with an empty quarter name
		{"Fifth Quarter", 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseQuarter(tc.in)
			if tc.err {
				if err == nil {
					t.Fatalf("want error, got %s", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got %s, want %s", got, tc.want)
			}
		})
	}
	if Annual.String() != "FY" || Q1.String() != "Q1" {
		t.Fatalf("canonical forms: %s %s", Annual, Q1)
	}
}

// The Shrawan 1 dates and the month lengths were checked by hand separately;
// each year's gap between Shrawan 1 dates must equal its nine remaining months
// plus the next year's first three.
func TestTableIsConsistent(t *testing.T) {
	for i := 0; i+1 < len(years); i++ {
		days := 0
		for _, n := range years[i].months[3:] {
			days += n
		}
		for _, n := range years[i+1].months[:3] {
			days += n
		}
		got := int(years[i+1].shrawan1.Sub(years[i].shrawan1).Hours() / 24)
		if got != days {
			t.Errorf("BS %d: Shrawan 1 gap is %d days, month lengths sum to %d", firstYear+i, got, days)
		}
		if m := years[i].shrawan1.Month(); m != time.July {
			t.Errorf("BS %d: Shrawan 1 is in %s, not July", firstYear+i, m)
		}
	}
}
