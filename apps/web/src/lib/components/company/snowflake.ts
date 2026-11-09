import type { Check } from '#lib/gen/ntx/v1/story_pb.js';

// The five axes of the snowflake, in drawing order from the top clockwise. The
// API names each check's axis, so the section chip and the axis read the same
// checks without the web keeping a list of names.
export const AXES = [
	{ id: 'value', label: 'Value' },
	{ id: 'past', label: 'Past' },
	{ id: 'health', label: 'Health' },
	{ id: 'dividend', label: 'Dividend' },
	{ id: 'ownership', label: 'Ownership' }
] as const;

export type AxisId = (typeof AXES)[number]['id'];

export function checksFor(all: Check[], axis: AxisId): Check[] {
	return all.filter((c) => c.axis === axis);
}

// The share of all checks that hold (ADR 0005): a check on watch or without
// history scores zero, so thin data draws a small axis. No weighting: the API's verdicts are the rule.
export function score(checks: Check[]): {
	value: number;
	holds: number;
	evaluated: number;
	total: number;
} {
	const total = checks.length;
	const evaluated = checks.filter((c) => c.verdict !== 'needs more history').length;
	const holds = checks.filter((c) => c.verdict === 'holds').length;
	return { value: total === 0 ? 0 : holds / total, holds, evaluated, total };
}
