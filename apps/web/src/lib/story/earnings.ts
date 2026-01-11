/**
 * Earnings Story Generator
 *
 * Generates narrative insights about a company's earnings trends.
 */

import type { Fundamental } from '$lib/gen/ntx/v1/common_pb';
import type { EarningsStory } from './types';

interface EarningsContext {
	current: Fundamental;
	history: Fundamental[];
}

/**
 * Sort fundamentals by fiscal year (most recent first)
 */
function sortByYear(fundamentals: Fundamental[]): Fundamental[] {
	return [...fundamentals].sort((a, b) => b.fiscalYear.localeCompare(a.fiscalYear));
}

/**
 * Calculate YoY EPS growth rate
 */
function calculateEpsGrowth(current: Fundamental, previous: Fundamental): number | undefined {
	const currentEps = current.eps;
	const previousEps = previous.eps;

	if (currentEps === undefined || previousEps === undefined || previousEps === 0) {
		return undefined;
	}

	return ((currentEps - previousEps) / Math.abs(previousEps)) * 100;
}

/**
 * Count consecutive years of EPS growth
 */
function countGrowthStreak(history: Fundamental[]): number {
	const sorted = sortByYear(history);
	let streak = 0;

	for (let i = 0; i < sorted.length - 1; i++) {
		const current = sorted[i].eps;
		const previous = sorted[i + 1].eps;

		if (current === undefined || previous === undefined) break;
		if (current <= previous) break;

		streak++;
	}

	return streak;
}

/**
 * Generate headline based on EPS growth
 */
function generateHeadline(epsGrowth: number | undefined): string {
	if (epsGrowth === undefined) {
		return 'Earnings data available';
	}

	if (epsGrowth > 20) {
		return 'Earnings are surging';
	}

	if (epsGrowth > 10) {
		return 'Earnings are growing strongly';
	}

	if (epsGrowth > 0) {
		return 'Earnings are growing modestly';
	}

	if (epsGrowth > -10) {
		return 'Earnings have dipped slightly';
	}

	if (epsGrowth > -20) {
		return 'Earnings are declining';
	}

	return 'Earnings have dropped significantly';
}

/**
 * Generate detail sentence with specific numbers
 */
function generateDetail(
	current: Fundamental,
	previous: Fundamental | undefined,
	epsGrowth: number | undefined
): string {
	const currentEps = current.eps;

	if (currentEps === undefined) {
		return 'EPS data not available for comparison.';
	}

	if (epsGrowth === undefined || previous === undefined) {
		return `Current EPS is Rs. ${currentEps.toFixed(2)}.`;
	}

	const direction = epsGrowth >= 0 ? 'increased' : 'decreased';
	return `EPS ${direction} ${Math.abs(epsGrowth).toFixed(0)}% to Rs. ${currentEps.toFixed(2)} from Rs. ${previous.eps?.toFixed(2) ?? 'N/A'}.`;
}

/**
 * Generate trend sentence for multi-year context
 */
function generateTrendSentence(streak: number, history: Fundamental[]): string {
	if (streak >= 4) {
		return `${streak} consecutive years of earnings growth.`;
	}

	if (streak >= 2) {
		return `${streak} years of consistent growth.`;
	}

	if (streak === 1) {
		return 'Earnings grew this year.';
	}

	// Check for turnaround
	const sorted = sortByYear(history);
	if (sorted.length >= 2) {
		const recent = sorted[0].eps ?? 0;
		const previous = sorted[1].eps ?? 0;
		if (recent > 0 && previous < 0) {
			return 'A turnaround story – back to profitability.';
		}
	}

	// Check for decline
	if (sorted.length >= 2) {
		const recent = sorted[0].eps ?? 0;
		const previous = sorted[1].eps ?? 0;
		if (recent < previous) {
			return 'Earnings declined from the previous year.';
		}
	}

	return 'Earnings have been variable.';
}

/**
 * Generate the complete earnings story
 */
export function generateEarningsStory(context: EarningsContext): EarningsStory {
	const { current, history } = context;
	const sorted = sortByYear(history);
	const previous = sorted.length > 1 ? sorted[1] : undefined;

	const epsGrowth = previous ? calculateEpsGrowth(current, previous) : undefined;
	const streak = countGrowthStreak(history);

	return {
		headline: generateHeadline(epsGrowth),
		detail: generateDetail(current, previous, epsGrowth),
		trendSentence: generateTrendSentence(streak, history),
		epsGrowth,
		growthStreak: streak
	};
}
