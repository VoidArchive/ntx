package worker

import (
	"database/sql"
	"math"
	"testing"
)

func TestNullHelpersTreatZeroAsNull(t *testing.T) {
	if got := nullString(""); got.Valid {
		t.Fatalf("nullString(\"\") = %+v, want invalid", got)
	}
	if got := nullString("x"); got != (sql.NullString{String: "x", Valid: true}) {
		t.Fatalf("nullString(\"x\") = %+v", got)
	}
	if got := nullFloat64(0); got.Valid {
		t.Fatalf("nullFloat64(0) = %+v, want invalid", got)
	}
	if got := nullFloat64(1.5); got != (sql.NullFloat64{Float64: 1.5, Valid: true}) {
		t.Fatalf("nullFloat64(1.5) = %+v", got)
	}
	if got := nullFloat64Ptr(nil); got.Valid {
		t.Fatalf("nullFloat64Ptr(nil) = %+v, want invalid", got)
	}
	// A pointer to zero is a real zero (e.g. a 0% right share), not a missing value.
	zero := 0.0
	if got := nullFloat64Ptr(&zero); got != (sql.NullFloat64{Float64: 0, Valid: true}) {
		t.Fatalf("nullFloat64Ptr(&0) = %+v", got)
	}
	if got := nullInt64(0); got.Valid {
		t.Fatalf("nullInt64(0) = %+v, want invalid", got)
	}
	if got := nullInt64(7); got != (sql.NullInt64{Int64: 7, Valid: true}) {
		t.Fatalf("nullInt64(7) = %+v", got)
	}
}

func TestSafeInt32(t *testing.T) {
	tests := []struct {
		in   int64
		want int32
	}{
		{0, 0},
		{2814, 2814},
		{math.MaxInt32, math.MaxInt32},
		{math.MaxInt32 + 1, math.MaxInt32},
		{math.MaxInt64, math.MaxInt32},
	}
	for _, tt := range tests {
		if got := safeInt32(tt.in); got != tt.want {
			t.Errorf("safeInt32(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
