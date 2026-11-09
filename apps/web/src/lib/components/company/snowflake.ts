import type { Check } from '#lib/gen/ntx/v1/story_pb.js';

// The five axes of the snowflake, in drawing order from the top clockwise. Each
// is also a page section, so the section chip and the axis read the same checks.
export const AXES = [
	{
		id: 'valuation',
		label: 'Value',
		checks: ['PE against its own history', 'PB against the sector']
	},
	{ id: 'past', label: 'Past', checks: ['Quarterly profit', 'EPS growth'] },
	{ id: 'health', label: 'Health', checks: ['Book value growth'] },
	{ id: 'dividend', label: 'Dividend', checks: ['Dividend record', 'Cash payout'] },
	{ id: 'ownership', label: 'Ownership', checks: ['Promoter holding'] }
] as const;

export type AxisId = (typeof AXES)[number]['id'];

export function checksFor(all: Check[], names: readonly string[]): Check[] {
	return all.filter((c) => names.includes(c.name));
}

// The share of evaluated checks that hold; a check without history is left out,
// and an axis with nothing evaluated scores 0. No weighting: the API's verdicts are the rule.
export function score(checks: Check[]): { value: number; holds: number; evaluated: number } {
	const evaluated = checks.filter((c) => c.verdict !== 'needs more history');
	const holds = evaluated.filter((c) => c.verdict === 'holds').length;
	if (evaluated.length === 0) return { value: 0, holds: 0, evaluated: 0 };
	return { value: holds / evaluated.length, holds, evaluated: evaluated.length };
}
