package worker

import "testing"

func TestTradedOn(t *testing.T) {
	tests := []struct {
		name         string
		asOf         string
		businessDate string
		want         bool
	}{
		{"closed today after session", "2026-11-04T15:00:00", "2026-11-04", true},
		{"still open today", "2026-11-04T11:00:00", "2026-11-04", true},
		{"holiday: status stuck on last session", "2026-11-02T15:00:00", "2026-11-04", false},
		{"empty asOf", "", "2026-11-04", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tradedOn(tt.asOf, tt.businessDate); got != tt.want {
				t.Fatalf("tradedOn(%q, %q) = %v, want %v", tt.asOf, tt.businessDate, got, tt.want)
			}
		})
	}
}
