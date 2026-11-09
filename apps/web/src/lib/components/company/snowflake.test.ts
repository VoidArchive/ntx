import { describe, expect, it } from 'vitest';
import type { Check } from '#lib/gen/ntx/v1/story_pb.js';
import { checksFor, score } from './snowflake.js';

// Checks as the API returns them: each names its axis (GLOSSARY: Check).
function check(name: string, axis: string, verdict: string): Check {
	return { name, rule: '', verdict, axis } as Check;
}

const checks = [
	check('PE against its own history', 'value', 'holds'),
	check('PB against the sector', 'value', 'watch'),
	check('PE against the sector', 'value', 'needs more history'),
	check('Dividend record', 'dividend', 'needs more history'),
	check('Cash payout', 'dividend', 'needs more history')
];

describe('checksFor', () => {
	it('groups by the axis the API names, not by check name', () => {
		expect(checksFor(checks, 'value').map((c) => c.name)).toEqual([
			'PE against its own history',
			'PB against the sector',
			'PE against the sector'
		]);
		expect(checksFor(checks, 'health')).toEqual([]);
	});
});

describe('score', () => {
	it('counts holds over every check, so a check without history scores zero', () => {
		expect(score(checksFor(checks, 'value'))).toEqual({
			value: 1 / 3,
			holds: 1,
			evaluated: 2,
			total: 3
		});
	});

	it('is zero, not undefined, when nothing on the axis evaluated', () => {
		expect(score(checksFor(checks, 'dividend'))).toEqual({
			value: 0,
			holds: 0,
			evaluated: 0,
			total: 2
		});
	});

	it('is zero for an axis with no checks at all', () => {
		expect(score([])).toEqual({ value: 0, holds: 0, evaluated: 0, total: 0 });
	});
});
