package market

import (
	"context"
	"testing"
	"time"
)

func TestIsOpenAt(t *testing.T) {
	m := New(nil) // No DB, local logic only
	ctx := context.Background()

	tests := []struct {
		name string
		time time.Time
		want bool
	}{
		{
			name: "Sunday 12:00 NPT - open",
			time: time.Date(2026, 1, 4, 12, 0, 0, 0, NPT), // Sunday
			want: true,
		},
		{
			name: "Sunday 10:00 NPT - before open",
			time: time.Date(2026, 1, 4, 10, 0, 0, 0, NPT),
			want: false,
		},
		{
			name: "Sunday 15:00 NPT - closed",
			time: time.Date(2026, 1, 4, 15, 0, 0, 0, NPT),
			want: false,
		},
		{
			name: "Sunday 11:00 NPT - exactly at open",
			time: time.Date(2026, 1, 4, 11, 0, 0, 0, NPT),
			want: true,
		},
		{
			name: "Sunday 14:59 NPT - just before close",
			time: time.Date(2026, 1, 4, 14, 59, 0, 0, NPT),
			want: true,
		},
		{
			name: "Friday 12:00 NPT - holiday",
			time: time.Date(2026, 1, 2, 12, 0, 0, 0, NPT), // Friday
			want: false,
		},
		{
			name: "Saturday 12:00 NPT - holiday",
			time: time.Date(2026, 1, 3, 12, 0, 0, 0, NPT), // Saturday
			want: false,
		},
		{
			name: "Monday 13:00 NPT - open",
			time: time.Date(2026, 1, 5, 13, 0, 0, 0, NPT), // Monday
			want: true,
		},
		{
			name: "Thursday 14:00 NPT - open",
			time: time.Date(2026, 1, 8, 14, 0, 0, 0, NPT), // Thursday
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.IsOpenAt(ctx, tt.time)
			if got != tt.want {
				t.Errorf("IsOpenAt(%v) = %v, want %v", tt.time, got, tt.want)
			}
		})
	}
}

func TestIsTradingDay(t *testing.T) {
	m := New(nil)
	ctx := context.Background()

	tests := []struct {
		name string
		time time.Time
		want bool
	}{
		{
			name: "Sunday - trading day",
			time: time.Date(2026, 1, 4, 12, 0, 0, 0, NPT),
			want: true,
		},
		{
			name: "Monday - trading day",
			time: time.Date(2026, 1, 5, 12, 0, 0, 0, NPT),
			want: true,
		},
		{
			name: "Tuesday - trading day",
			time: time.Date(2026, 1, 6, 12, 0, 0, 0, NPT),
			want: true,
		},
		{
			name: "Wednesday - trading day",
			time: time.Date(2026, 1, 7, 12, 0, 0, 0, NPT),
			want: true,
		},
		{
			name: "Thursday - trading day",
			time: time.Date(2026, 1, 8, 12, 0, 0, 0, NPT),
			want: true,
		},
		{
			name: "Friday - not trading day",
			time: time.Date(2026, 1, 2, 12, 0, 0, 0, NPT),
			want: false,
		},
		{
			name: "Saturday - not trading day",
			time: time.Date(2026, 1, 3, 12, 0, 0, 0, NPT),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.IsTradingDay(ctx, tt.time)
			if got != tt.want {
				t.Errorf("IsTradingDay(%v) = %v, want %v (weekday: %s)", tt.time, got, tt.want, tt.time.Weekday())
			}
		})
	}
}

func TestNextOpenFrom(t *testing.T) {
	m := New(nil)
	ctx := context.Background()

	tests := []struct {
		name string
		from time.Time
		want time.Time
	}{
		{
			name: "Sunday 10:00 - same day",
			from: time.Date(2026, 1, 4, 10, 0, 0, 0, NPT),
			want: time.Date(2026, 1, 4, 11, 0, 0, 0, NPT),
		},
		{
			name: "Sunday 12:00 - next day (Monday)",
			from: time.Date(2026, 1, 4, 12, 0, 0, 0, NPT),
			want: time.Date(2026, 1, 5, 11, 0, 0, 0, NPT),
		},
		{
			name: "Thursday 16:00 - skip Fri/Sat to Sunday",
			from: time.Date(2026, 1, 8, 16, 0, 0, 0, NPT),
			want: time.Date(2026, 1, 11, 11, 0, 0, 0, NPT),
		},
		{
			name: "Friday 10:00 - skip to Sunday",
			from: time.Date(2026, 1, 2, 10, 0, 0, 0, NPT),
			want: time.Date(2026, 1, 4, 11, 0, 0, 0, NPT),
		},
		{
			name: "Saturday 10:00 - skip to Sunday",
			from: time.Date(2026, 1, 3, 10, 0, 0, 0, NPT),
			want: time.Date(2026, 1, 4, 11, 0, 0, 0, NPT),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.NextOpenFrom(ctx, tt.from)
			if !got.Equal(tt.want) {
				t.Errorf("NextOpenFrom(%v) = %v, want %v", tt.from, got, tt.want)
			}
		})
	}
}

func TestNextCloseFrom(t *testing.T) {
	m := New(nil)
	ctx := context.Background()

	tests := []struct {
		name string
		from time.Time
		want time.Time
	}{
		{
			name: "Sunday 12:00 - same day",
			from: time.Date(2026, 1, 4, 12, 0, 0, 0, NPT),
			want: time.Date(2026, 1, 4, 15, 0, 0, 0, NPT),
		},
		{
			name: "Sunday 16:00 - next day (Monday)",
			from: time.Date(2026, 1, 4, 16, 0, 0, 0, NPT),
			want: time.Date(2026, 1, 5, 15, 0, 0, 0, NPT),
		},
		{
			name: "Thursday 16:00 - skip Fri/Sat to Sunday",
			from: time.Date(2026, 1, 8, 16, 0, 0, 0, NPT),
			want: time.Date(2026, 1, 11, 15, 0, 0, 0, NPT),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.NextCloseFrom(ctx, tt.from)
			if !got.Equal(tt.want) {
				t.Errorf("NextCloseFrom(%v) = %v, want %v", tt.from, got, tt.want)
			}
		})
	}
}

func TestStatus(t *testing.T) {
	m := New(nil)
	ctx := context.Background()

	openTime := time.Date(2026, 1, 4, 12, 0, 0, 0, NPT) // Sunday 12:00
	status := m.StatusAt(ctx, openTime)

	if !status.IsOpen {
		t.Error("expected market to be open")
	}
	if status.State != StateOpen {
		t.Errorf("expected state %q, got %q", StateOpen, status.State)
	}

	closedTime := time.Date(2026, 1, 4, 16, 0, 0, 0, NPT) // Sunday 16:00
	status = m.StatusAt(ctx, closedTime)

	if status.IsOpen {
		t.Error("expected market to be closed")
	}
	if status.State != StateClosed {
		t.Errorf("expected state %q, got %q", StateClosed, status.State)
	}
}

func TestNPTTimezone(t *testing.T) {
	// Verify NPT is UTC+5:45
	_, offset := time.Now().In(NPT).Zone()
	expected := 5*60*60 + 45*60 // 5 hours 45 minutes in seconds

	if offset != expected {
		t.Errorf("NPT offset = %d, want %d", offset, expected)
	}
}
