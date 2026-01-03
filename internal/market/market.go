// Package market provides market hours logic for NEPSE.
//
// NEPSE operates 11:00-15:00 NPT, Sunday-Thursday.
// Friday and Saturday are holidays.
package market

import (
	"context"
	"errors"
	"time"

	"github.com/voidarchive/ntx/internal/database/sqlc"
)

// NPT is Nepal Time (UTC+5:45).
var NPT = time.FixedZone("NPT", 5*60*60+45*60)

const (
	// Market hours in NPT
	OpenHour  = 11
	CloseHour = 15

	// Date format for trading_days table
	DateFormat = "2006-01-02"
)

// State represents the current market state.
type State string

const (
	StateOpen   State = "open"
	StateClosed State = "closed"
)

// Status represents the current market status.
type Status struct {
	IsOpen bool
	State  State
	AsOf   time.Time
}

// Market provides market hours logic.
type Market struct {
	queries *sqlc.Queries
}

// New creates a new Market instance.
// If queries is nil, holiday lookups will be skipped (local logic only).
func New(queries *sqlc.Queries) *Market {
	return &Market{queries: queries}
}

// IsOpen returns true if the market is currently open.
// Uses local time-based logic: 11:00-15:00 NPT, Sunday-Thursday.
// If a database is configured, also checks for holidays.
func (m *Market) IsOpen(ctx context.Context) bool {
	return m.IsOpenAt(ctx, time.Now())
}

// IsOpenAt returns true if the market would be open at the given time.
func (m *Market) IsOpenAt(ctx context.Context, t time.Time) bool {
	npt := t.In(NPT)

	if !m.IsTradingDay(ctx, npt) {
		return false
	}

	hour := npt.Hour()
	return hour >= OpenHour && hour < CloseHour
}

// IsTradingDay returns true if the given date is a trading day.
// Checks: not Friday/Saturday, and not a known holiday.
func (m *Market) IsTradingDay(ctx context.Context, t time.Time) bool {
	npt := t.In(NPT)
	weekday := npt.Weekday()

	// Friday and Saturday are holidays
	if weekday == time.Friday || weekday == time.Saturday {
		return false
	}

	// Check database for holidays if available
	if m.queries != nil {
		date := npt.Format(DateFormat)
		day, err := m.queries.GetTradingDay(ctx, date)
		if err == nil {
			return day.IsOpen == 1
		}
		// If not found in DB, assume it's a trading day
		// Other errors are ignored and we fall back to default logic
	}

	return true
}

// Status returns the current market status.
func (m *Market) Status(ctx context.Context) Status {
	return m.StatusAt(ctx, time.Now())
}

// StatusAt returns the market status at the given time.
func (m *Market) StatusAt(ctx context.Context, t time.Time) Status {
	isOpen := m.IsOpenAt(ctx, t)
	state := StateClosed
	if isOpen {
		state = StateOpen
	}

	return Status{
		IsOpen: isOpen,
		State:  state,
		AsOf:   t,
	}
}

// NextOpen returns the next time the market will open.
// If the market is currently open, returns the current day's open time.
func (m *Market) NextOpen(ctx context.Context) time.Time {
	return m.NextOpenFrom(ctx, time.Now())
}

// NextOpenFrom returns the next market open time from the given time.
func (m *Market) NextOpenFrom(ctx context.Context, t time.Time) time.Time {
	npt := t.In(NPT)

	// Start from today's open time
	open := time.Date(npt.Year(), npt.Month(), npt.Day(), OpenHour, 0, 0, 0, NPT)

	// If we're past today's open, start from tomorrow
	if npt.Hour() >= OpenHour {
		open = open.AddDate(0, 0, 1)
	}

	// Find next trading day (max 10 days to avoid infinite loop)
	for i := 0; i < 10; i++ {
		if m.IsTradingDay(ctx, open) {
			return open
		}
		open = open.AddDate(0, 0, 1)
	}

	return open
}

// NextClose returns the next time the market will close.
// If the market is currently closed, returns the next trading day's close time.
func (m *Market) NextClose(ctx context.Context) time.Time {
	return m.NextCloseFrom(ctx, time.Now())
}

// NextCloseFrom returns the next market close time from the given time.
func (m *Market) NextCloseFrom(ctx context.Context, t time.Time) time.Time {
	npt := t.In(NPT)

	// Start from today's close time
	close := time.Date(npt.Year(), npt.Month(), npt.Day(), CloseHour, 0, 0, 0, NPT)

	// If we're past today's close, start from tomorrow
	if npt.Hour() >= CloseHour {
		close = close.AddDate(0, 0, 1)
	}

	// Find next trading day (max 10 days to avoid infinite loop)
	for i := 0; i < 10; i++ {
		if m.IsTradingDay(ctx, close) {
			return close
		}
		close = close.AddDate(0, 0, 1)
	}

	return close
}

// UntilOpen returns the duration until the market opens.
// Returns 0 if the market is currently open.
func (m *Market) UntilOpen(ctx context.Context) time.Duration {
	if m.IsOpen(ctx) {
		return 0
	}
	return time.Until(m.NextOpen(ctx))
}

// UntilClose returns the duration until the market closes.
// Returns 0 if the market is currently closed.
func (m *Market) UntilClose(ctx context.Context) time.Duration {
	if !m.IsOpen(ctx) {
		return 0
	}
	return time.Until(m.NextClose(ctx))
}

// RecordTradingDay records whether a date was a trading day.
// Called by background worker after syncing from go-nepse.
func (m *Market) RecordTradingDay(ctx context.Context, date time.Time, isOpen bool, status string) error {
	if m.queries == nil {
		return errors.New("no database configured")
	}

	isOpenInt := int64(0)
	if isOpen {
		isOpenInt = 1
	}

	return m.queries.UpsertTradingDay(ctx, sqlc.UpsertTradingDayParams{
		Date:   date.In(NPT).Format(DateFormat),
		IsOpen: isOpenInt,
		Status: status,
	})
}
