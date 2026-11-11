package portfolio

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"

	ntxv1 "github.com/voidarchive/ntx/gen/go/ntx/v1"
	"github.com/voidarchive/ntx/internal/database/sqlc"
	"github.com/voidarchive/ntx/internal/fiscal"
)

// rows is what one company's metric gates read (ADR 0001: evaluated in the request, nothing stored).
type rows struct {
	fundamentals []sqlc.Fundamental
	actions      []sqlc.CorporateAction // newest first
	ownership    []sqlc.OwnershipHistory
	bank         []sqlc.BankMetric // owner-entered; the exchange publishes none
}

// judge evaluates every metric gate of every company holding against the latest rows.
func (s *Service) judge(ctx context.Context, holdings []*ntxv1.Holding) error {
	for _, h := range holdings {
		if len(h.GetGates()) == 0 || h.GetIsFund() {
			continue
		}
		c, err := s.queries.GetCompany(ctx, h.GetSymbol())
		if err != nil {
			return fmt.Errorf("company %s: %w", h.GetSymbol(), err)
		}
		var r rows
		if r.fundamentals, err = s.queries.ListFundamentalsByCompany(ctx, c.ID); err != nil {
			return err
		}
		if r.actions, err = s.queries.ListCorporateActionsByCompany(ctx, c.ID); err != nil {
			return err
		}
		if r.ownership, err = s.queries.ListOwnershipHistory(ctx, c.ID); err != nil {
			return err
		}
		if r.bank, err = s.queries.ListBankMetricsByCompany(ctx, c.ID); err != nil {
			return err
		}
		for _, g := range h.GetGates() {
			if g.GetKind() != ntxv1.GateKind_GATE_KIND_METRIC {
				continue
			}
			// The drawdown gate runs only where the owner switched it on; off is not missing data.
			if g.GetMetric() == "drawdown_percent" && !h.GetDrawdownGate() {
				continue
			}
			v := evaluate(
				directionOf(g.GetDirection()),
				g.GetThreshold(),
				int(g.GetPeriods()),
				series(g.GetMetric(), h, r),
			)
			g.Verdict, g.Value, g.Prior, g.Period, g.Source = v.Verdict, v.Value, v.Prior, v.Period, v.Source
		}
	}
	return nil
}

// series is a metric's figures oldest first, each with the period it was filed for.
func series(metric string, h *ntxv1.Holding, r rows) []point {
	switch metric {
	case "eps", "pe", "roe":
		return quarterly(r.fundamentals, metric)
	case "npl_percent", "car_percent", "loan_growth_percent", "distributable_profit":
		return bankSeries(r.fundamentals, r.bank, metric)
	case "book_value":
		return yearly(r.fundamentals)
	case "cash_dividend_percent", "bonus_percent":
		return notices(r.actions, metric)
	case "promoter_percent":
		var out []point
		for _, o := range r.ownership {
			if o.PromoterPercent.Valid {
				out = append(out, point{Value: o.PromoterPercent.Float64, Period: o.AsOf})
			}
		}
		return out
	case "weight_percent":
		if h.GetQuantity() > 0 && h.Close != nil {
			return []point{{Value: h.GetWeightPercent(), Period: h.GetCloseDay()}}
		}
	case "drawdown_percent":
		// Against the bonus-adjusted WAC.
		if h.GetQuantity() > 0 && h.GetWac() > 0 && h.Close != nil {
			return []point{{Value: (h.GetWac() - h.GetClose()) / h.GetWac() * 100, Period: h.GetCloseDay()}}
		}
	}
	return nil
}

type filing struct {
	year    fiscal.Year
	quarter fiscal.Quarter
	row     sqlc.Fundamental
}

func (f filing) rank() int { return int(f.year)*10 + int(f.quarter) }

// period is the quarter as a verdict names it: "Q1 2083/84".
func (f filing) period() string { return f.quarter.String() + " " + f.year.String() }

func filings(rows []sqlc.Fundamental) []filing {
	var out []filing
	for _, r := range rows {
		y, err := fiscal.ParseFilingYear(r.FiscalYearNp.String, r.FiscalYear)
		if err != nil {
			continue
		}
		q, err := fiscal.ParseQuarter(r.Quarter)
		if err != nil {
			continue
		}
		out = append(out, filing{y, q, r})
	}
	slices.SortFunc(out, func(a, b filing) int { return cmp.Compare(a.rank(), b.rank()) })
	return out
}

func quarterly(rows []sqlc.Fundamental, metric string) []point {
	var out []point
	for _, f := range filings(rows) {
		if f.quarter == fiscal.Annual {
			continue
		}
		var v sql.NullFloat64
		switch metric {
		case "eps":
			v = f.row.Eps
		case "pe":
			v = f.row.PeRatio
		case "roe":
			if f.row.Eps.Valid && f.row.BookValue.Valid && f.row.BookValue.Float64 > 0 {
				v = sql.NullFloat64{Float64: f.row.Eps.Float64 * 100 / f.row.BookValue.Float64, Valid: true}
			}
		}
		if v.Valid {
			out = append(
				out,
				point{
					Value:  v.Float64,
					Period: f.period(),
					Source: f.row.SourceUrl.String,
				},
			)
		}
	}
	return out
}

// yearly reads book value from the annual report, else the year's Q4 filing.
func yearly(rows []sqlc.Fundamental) []point {
	byYear := map[fiscal.Year]filing{}
	for _, f := range filings(rows) {
		if f.quarter != fiscal.Annual && f.quarter != fiscal.Q4 {
			continue
		}
		if prev, ok := byYear[f.year]; ok && prev.quarter == fiscal.Annual {
			continue
		}
		byYear[f.year] = f
	}
	var out []point
	for _, y := range slices.Sorted(maps.Keys(byYear)) {
		if f := byYear[y]; f.row.BookValue.Valid {
			out = append(
				out,
				point{Value: f.row.BookValue.Float64, Period: "FY " + y.String(), Source: f.row.SourceUrl.String},
			)
		}
	}
	return out
}

// notices keeps the newest notice per fiscal year, as the story engine does.
func notices(actions []sqlc.CorporateAction, metric string) []point {
	byYear := map[fiscal.Year]sqlc.CorporateAction{}
	for _, a := range actions {
		y, err := fiscal.ParseYear(a.FiscalYear)
		if err != nil {
			continue
		}
		if _, ok := byYear[y]; !ok {
			byYear[y] = a
		}
	}
	var out []point
	for _, y := range slices.Sorted(maps.Keys(byYear)) {
		a := byYear[y]
		v := a.CashDividend
		if metric == "bonus_percent" {
			v = a.BonusPercentage
		}
		// A notice silent on this figure is not a zero.
		if v.Valid {
			out = append(out, point{Value: v.Float64, Period: "FY " + y.String(), Source: a.NoticeUrl.String})
		}
	}
	return out
}

// caps reads every holding against the single, sector, position and basin rules.
// Funds sit in the sector sums but carry no single-holding cap.
func caps(p *ntxv1.GetPortfolioResponse, r sqlc.Rule, plants []sqlc.HydroPlant) []*ntxv1.Cap {
	var out []*ntxv1.Cap
	sectors := map[string]float64{}
	for _, h := range p.GetHoldings() {
		if h.GetQuantity() == 0 || h.Close == nil {
			continue
		}
		sectors[h.GetSector()] += h.GetWeightPercent()
		if h.GetIsFund() {
			continue
		}
		h.CapVerdict = Holds
		detail := func(verdict, name string, cap float64) {
			h.CapVerdict = verdict
			out = append(out, &ntxv1.Cap{Rule: "Single holding", Verdict: verdict, Detail: fmt.Sprintf(
				"%s %.1f%% of the %s%% %s cap", h.GetSymbol(), h.GetWeightPercent(), num(cap), name,
			)})
		}
		switch {
		case h.GetWeightPercent() > r.HardCapPercent:
			detail(Triggered, "hard", r.HardCapPercent)
		case h.GetWeightPercent() > r.SoftCapPercent:
			detail(Watch, "soft", r.SoftCapPercent)
		}
	}

	out = append(out, sectorCap(sectors, r.SectorCapPercent))

	positions := &ntxv1.Cap{
		Rule:    "Positions",
		Verdict: Holds,
		Detail:  fmt.Sprintf("%d of %d", p.GetPositions(), r.MaxPositions),
	}
	if int64(p.GetPositions()) > r.MaxPositions {
		positions.Verdict = Triggered
	}
	out = append(out, positions)

	if r.OneHydroPerBasin != 0 {
		out = append(out, basinCap(p.GetHoldings(), plants))
	}
	return out
}

// sectorCap names every sector over the cap, else the largest one.
func sectorCap(sectors map[string]float64, cap float64) *ntxv1.Cap {
	c := &ntxv1.Cap{Rule: "Sector cap", Verdict: Holds, Detail: "no sector held"}
	names := slices.Sorted(maps.Keys(sectors))
	if len(names) == 0 {
		return c
	}
	figure := func(name string) string {
		return fmt.Sprintf("%s %.1f%% of the %s%% sector cap", name, sectors[name], num(cap))
	}
	var over []string
	for _, name := range names {
		if sectors[name] > cap {
			over = append(over, figure(name))
		}
	}
	if len(over) > 0 {
		c.Verdict, c.Detail = Triggered, strings.Join(over, ", ")
		return c
	}
	top := slices.MaxFunc(names, func(a, b string) int { return cmp.Compare(sectors[a], sectors[b]) })
	c.Detail = figure(top)
	return c
}

// basinCap finds basins where two held hydro companies both have a plant (GLOSSARY: Basin).
func basinCap(holdings []*ntxv1.Holding, plants []sqlc.HydroPlant) *ntxv1.Cap {
	held := map[string]bool{}
	for _, h := range holdings {
		if h.GetQuantity() > 0 && !h.GetIsFund() {
			held[h.GetSymbol()] = true
		}
	}
	basins := map[string][]string{}
	for _, pl := range plants {
		if held[pl.Symbol] && !slices.Contains(basins[pl.Basin], pl.Symbol) {
			basins[pl.Basin] = append(basins[pl.Basin], pl.Symbol)
		}
	}
	var shared []string
	for _, b := range slices.Sorted(maps.Keys(basins)) {
		if len(basins[b]) > 1 {
			shared = append(shared, fmt.Sprintf("%s basin: %s", b, list(basins[b])))
		}
	}
	c := &ntxv1.Cap{Rule: "One hydro per basin", Verdict: Holds, Detail: "no basin held twice"}
	if len(shared) > 0 {
		c.Verdict, c.Detail = Triggered, strings.Join(shared, "; ")
	}
	return c
}

// list joins names as prose: "MEN", "MEN and SAHAS", "MEN, API and SAHAS".
func list(names []string) string {
	if len(names) < 2 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}

func num(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func rulesToProto(r sqlc.Rule) *ntxv1.Rules {
	return &ntxv1.Rules{
		SoftCapPercent:          r.SoftCapPercent,
		HardCapPercent:          r.HardCapPercent,
		SectorCapPercent:        r.SectorCapPercent,
		MaxPositions:            int32(min(r.MaxPositions, 1<<31-1)), //nolint:gosec // clamped
		MinEntryRs:              r.MinEntryRs,
		OneHydroPerBasin:        r.OneHydroPerBasin != 0,
		CandidateMinRating:      r.CandidateMinRating,
		CandidateDividendRecord: r.CandidateDividendRecord != 0,
		MonthlyBudgetRs:         r.MonthlyBudgetRs,
		MultiplierBelowMedian:   r.MultiplierBelowMedian,
		MultiplierBottomQuarter: r.MultiplierBottomQuarter,
	}
}

func flag(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// validRules is the one guard on the rules row, for the form and the import alike.
func validRules(soft, hard float64, positions int64) error {
	if soft <= 0 || hard < soft || positions < 1 {
		return errors.New("the soft cap must be positive, the hard cap at least the soft cap, and positions at least 1")
	}
	return nil
}

func (s *Service) UpdateRules(
	ctx context.Context,
	req *connect.Request[ntxv1.UpdateRulesRequest],
) (*connect.Response[ntxv1.UpdateRulesResponse], error) {
	r := req.Msg.GetRules()
	if err := validRules(r.GetSoftCapPercent(), r.GetHardCapPercent(), int64(r.GetMaxPositions())); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	row, err := s.queries.UpdateRules(ctx, sqlc.UpdateRulesParams{
		SoftCapPercent:          r.GetSoftCapPercent(),
		HardCapPercent:          r.GetHardCapPercent(),
		SectorCapPercent:        r.GetSectorCapPercent(),
		MaxPositions:            int64(r.GetMaxPositions()),
		MinEntryRs:              r.GetMinEntryRs(),
		OneHydroPerBasin:        flag(r.GetOneHydroPerBasin()),
		CandidateMinRating:      strings.ToUpper(strings.TrimSpace(r.GetCandidateMinRating())),
		CandidateDividendRecord: flag(r.GetCandidateDividendRecord()),
		MonthlyBudgetRs:         r.GetMonthlyBudgetRs(),
		MultiplierBelowMedian:   r.GetMultiplierBelowMedian(),
		MultiplierBottomQuarter: r.GetMultiplierBottomQuarter(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&ntxv1.UpdateRulesResponse{Rules: rulesToProto(row)}), nil
}

func (s *Service) UpsertGate(
	ctx context.Context,
	req *connect.Request[ntxv1.UpsertGateRequest],
) (*connect.Response[ntxv1.UpsertGateResponse], error) {
	m := req.Msg
	h, err := s.holding(ctx, m.GetSymbol())
	if err != nil {
		return nil, err
	}
	if err := s.validGate(ctx, h.Symbol, m); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	periods := max(int64(m.GetPeriods()), 1)
	var row sqlc.Gate
	if m.GetId() == 0 {
		row, err = s.queries.InsertGate(ctx, sqlc.InsertGateParams{
			HoldingID: h.ID,
			Kind:      gateKindOf(m.GetKind()),
			Metric:    m.GetMetric(),
			Direction: directionOf(m.GetDirection()),
			Threshold: m.GetThreshold(),
			Periods:   periods,
			Text:      strings.TrimSpace(m.GetText()),
			CheckedOn: m.GetCheckedOn(),
		})
	} else {
		row, err = s.queries.UpdateGate(ctx, sqlc.UpdateGateParams{
			ID:        m.GetId(),
			HoldingID: h.ID,
			Kind:      gateKindOf(m.GetKind()),
			Metric:    m.GetMetric(),
			Direction: directionOf(m.GetDirection()),
			Threshold: m.GetThreshold(),
			Periods:   periods,
			Text:      strings.TrimSpace(m.GetText()),
			CheckedOn: m.GetCheckedOn(),
		})
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no gate %d on %s", m.GetId(), h.Symbol))
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	p, err := s.portfolio(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, x := range p.GetHoldings() {
		for _, g := range x.GetGates() {
			if g.GetId() == row.ID {
				return connect.NewResponse(&ntxv1.UpsertGateResponse{Gate: g}), nil
			}
		}
	}
	return nil, connect.NewError(connect.CodeInternal, errors.New("gate vanished after upsert"))
}

// validGate refuses what the evaluator could not read: an unknown metric, a
// metric gate without a direction, an event with no words, or a gate on a fund.
func (s *Service) validGate(ctx context.Context, symbol string, m *ntxv1.UpsertGateRequest) error {
	if m.GetCheckedOn() != "" {
		if _, err := time.Parse(time.DateOnly, m.GetCheckedOn()); err != nil {
			return err
		}
	}
	switch m.GetKind() {
	case ntxv1.GateKind_GATE_KIND_EVENT:
		if strings.TrimSpace(m.GetText()) == "" {
			return errors.New("an event gate needs its text")
		}
		return nil
	case ntxv1.GateKind_GATE_KIND_METRIC:
		if _, ok := metrics[m.GetMetric()]; !ok {
			return fmt.Errorf("unknown metric %q", m.GetMetric())
		}
		if m.GetDirection() == ntxv1.GateDirection_GATE_DIRECTION_UNSPECIFIED {
			return errors.New("a metric gate needs a direction")
		}
		if _, err := s.queries.GetCompany(ctx, symbol); errors.Is(err, sql.ErrNoRows) {
			return errors.New("gates read filings; a fund has none")
		} else if err != nil {
			return err
		}
		return nil
	}
	return errors.New("kind is required")
}

func (s *Service) DeleteGate(
	ctx context.Context,
	req *connect.Request[ntxv1.DeleteGateRequest],
) (*connect.Response[ntxv1.DeleteGateResponse], error) {
	if err := s.queries.DeleteGate(ctx, req.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&ntxv1.DeleteGateResponse{}), nil
}

func (s *Service) GetHoldingStory(
	ctx context.Context,
	req *connect.Request[ntxv1.GetHoldingStoryRequest],
) (*connect.Response[ntxv1.GetHoldingStoryResponse], error) {
	symbol := normSymbol(req.Msg.GetSymbol())
	p, err := s.portfolio(ctx, req.Msg.GetFundNavs())
	if err != nil {
		return nil, err
	}
	i := slices.IndexFunc(p.GetHoldings(), func(h *ntxv1.Holding) bool { return h.GetSymbol() == symbol })
	if i < 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no holding %q", symbol))
	}
	h := p.GetHoldings()[i]
	res := &ntxv1.GetHoldingStoryResponse{Holding: h}
	if res.BonusHint, err = s.bonusHint(ctx, h); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if res.BankMetrics, err = s.bankRows(ctx, h); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(res), nil
}

// bonusHint names a published bonus no bonus transaction on or after its notice credits.
// A hint only: nothing is derived from the notice (spec, section 2).
func (s *Service) bonusHint(ctx context.Context, h *ntxv1.Holding) (string, error) {
	if h.GetIsFund() {
		return "", nil
	}
	c, err := s.queries.GetCompany(ctx, h.GetSymbol())
	if err != nil {
		return "", err
	}
	a, err := s.queries.GetLatestCorporateActionByCompany(ctx, c.ID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && a.BonusPercentage.Float64 <= 0) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	for _, tx := range h.GetTransactions() {
		if tx.GetKind() == ntxv1.TransactionKind_TRANSACTION_KIND_BONUS && tx.GetDay() >= a.SubmittedDate.String {
			return "", nil
		}
	}
	fy := a.FiscalYear
	if y, err := fiscal.ParseYear(a.FiscalYear); err == nil {
		fy = y.String()
	}
	published := ""
	if a.SubmittedDate.String != "" {
		published = ", published " + a.SubmittedDate.String
	}
	return fmt.Sprintf(
		"A %s%% bonus for FY %s%s, is not yet credited in the book.",
		num(a.BonusPercentage.Float64), fy, published,
	), nil
}
