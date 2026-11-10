import { GateDirection, GateKind, type Gate } from '#lib/gen/ntx/v1/portfolio_pb.js';

// Display words for the API's verdicts (GLOSSARY: Triggered / Holds / Needs data).
export const VERDICT: Record<string, string> = {
	triggered: 'Triggered',
	holds: 'Holds',
	watch: 'Watch',
	'needs data': 'Needs data'
};

// Triggered and Needs data first (spec, section 5); event gates carry no verdict.
const RANK: Record<string, number> = { triggered: 0, 'needs data': 1, watch: 2, holds: 3 };
export const byVerdict = (a: Gate, b: Gate) => (RANK[a.verdict] ?? 4) - (RANK[b.verdict] ?? 4);

export const metrics = [
	{ value: 'eps', label: 'EPS' },
	{ value: 'pe', label: 'PE' },
	{ value: 'roe', label: 'ROE' },
	{ value: 'book_value', label: 'Book value' },
	{ value: 'cash_dividend_percent', label: 'Cash dividend %' },
	{ value: 'bonus_percent', label: 'Bonus %' },
	{ value: 'promoter_percent', label: 'Promoter holding %' },
	{ value: 'weight_percent', label: 'Weight %' },
	{ value: 'drawdown_percent', label: 'Drawdown %' },
	{ value: 'npl_percent', label: 'NPL %' },
	{ value: 'car_percent', label: 'CAR %' },
	{ value: 'loan_growth_percent', label: 'Loan growth %' },
	{ value: 'distributable_profit', label: 'Distributable profit' }
];

export const directions = [
	{ value: GateDirection.ABOVE, label: 'above' },
	{ value: GateDirection.BELOW, label: 'below' },
	{ value: GateDirection.NOT_BELOW_PRIOR, label: 'not below the prior period' },
	{ value: GateDirection.NOT_ABOVE_PRIOR, label: 'not above the prior period' }
];

export const isEvent = (g: Gate) => g.kind === GateKind.EVENT;

export function figure(v: number | undefined, unit: string) {
	if (v === undefined) return 'not recorded';
	const n = v.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	return unit === 'Rs' ? `Rs ${n}` : `${n}${unit}`;
}
