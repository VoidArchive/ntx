import { describe, expect, it } from 'vitest';
import { at } from './utils.js';

describe('at', () => {
	// A bank sector: PE from 11.65 to 111.48 with the median at 18.52.
	const range = { low: 11.65, median: 18.52, high: 111.48 };

	it('pins the median at the centre whatever the skew', () => {
		expect(at(range, 18.52)).toBe(50);
	});

	it('spreads the half below the median over the left half', () => {
		expect(at(range, 11.65)).toBe(0);
		expect(at(range, 16.29)).toBeCloseTo(33.8, 1);
	});

	it('spreads the half above the median over the right half', () => {
		expect(at(range, 111.48)).toBe(100);
		expect(at(range, 65)).toBeCloseTo(75, 1);
	});

	it('clamps outside the range and sits at the centre when a half is flat', () => {
		expect(at(range, 5)).toBe(0);
		expect(at(range, 500)).toBe(100);
		expect(at({ low: 2, median: 2, high: 4 }, 2)).toBe(50);
	});
});
