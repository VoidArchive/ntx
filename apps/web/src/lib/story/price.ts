/**
 * Price Story Generator
 *
 * Generates narrative insights about a stock's price position, trend, and volume.
 */

import type { Price } from '$lib/gen/ntx/v1/common_pb';
import type { PriceStory } from './types';

interface PriceContext {
	current: Price;
	history: Price[];
}

/**
 * Calculate 52-week high and low from price history
 */
function get52WeekRange(history: Price[]): { high: number; low: number } {
	if (history.length === 0) {
		return { high: 0, low: 0 };
	}

	let high = 0;
	let low = Infinity;

	for (const p of history) {
		const h = p.high ?? p.ltp ?? 0;
		const l = p.low ?? p.ltp ?? 0;
		if (h > high) high = h;
		if (l < low && l > 0) low = l;
	}

	return { high, low: low === Infinity ? 0 : low };
}

/**
 * Calculate position in 52-week range (0 = at low, 1 = at high)
 */
function calculateRangePosition(current: number, high: number, low: number): number {
	if (high === low) return 0.5;
	return (current - low) / (high - low);
}

/**
 * Generate position sentence based on where price is in 52W range
 */
function generatePositionSentence(currentPrice: number, high: number, low: number): string {
	const position = calculateRangePosition(currentPrice, high, low);

	const pctFromHigh = ((high - currentPrice) / high) * 100;
	const pctFromLow = ((currentPrice - low) / low) * 100;

	if (position > 0.9) {
		return `Trading near its 52-week high of Rs. ${high.toLocaleString()}`;
	}

	if (position < 0.1) {
		return `Trading near its 52-week low of Rs. ${low.toLocaleString()}`;
	}

	if (position > 0.5) {
		return `Trading ${pctFromHigh.toFixed(0)}% below its 52-week high of Rs. ${high.toLocaleString()}`;
	}

	return `Trading ${pctFromLow.toFixed(0)}% above its 52-week low of Rs. ${low.toLocaleString()}`;
}

/**
 * Calculate trend over a period (returns percentage change)
 */
function calculateTrend(history: Price[], days: number): number | null {
	if (history.length < 2) return null;

	// Sort by date descending (most recent first)
	const sorted = [...history].sort((a, b) => b.businessDate.localeCompare(a.businessDate));

	const recent = sorted[0];
	const lookbackIndex = Math.min(days, sorted.length - 1);
	const past = sorted[lookbackIndex];

	const recentPrice = recent.ltp ?? recent.close ?? 0;
	const pastPrice = past.ltp ?? past.close ?? 0;

	if (pastPrice === 0) return null;

	return ((recentPrice - pastPrice) / pastPrice) * 100;
}

/**
 * Generate trend sentence based on recent price movement
 */
function generateTrendSentence(history: Price[]): string {
	// Try 30-day trend first
	const monthTrend = calculateTrend(history, 30);
	const weekTrend = calculateTrend(history, 7);

	if (monthTrend !== null && Math.abs(monthTrend) > 5) {
		const direction = monthTrend > 0 ? 'up' : 'down';
		return `${direction.charAt(0).toUpperCase() + direction.slice(1)} ${Math.abs(monthTrend).toFixed(0)}% over the past month`;
	}

	if (weekTrend !== null && Math.abs(weekTrend) > 2) {
		const direction = weekTrend > 0 ? 'up' : 'down';
		return `${direction.charAt(0).toUpperCase() + direction.slice(1)} ${Math.abs(weekTrend).toFixed(0)}% this week`;
	}

	return 'Trading relatively flat recently';
}

/**
 * Calculate average volume from history
 */
function calculateAverageVolume(history: Price[]): number {
	const volumes = history.map((p) => Number(p.volume ?? 0)).filter((v) => v > 0);

	if (volumes.length === 0) return 0;
	return volumes.reduce((sum, v) => sum + v, 0) / volumes.length;
}

/**
 * Generate volume context sentence
 */
function generateVolumeContext(current: Price, history: Price[]): string | undefined {
	const currentVolume = Number(current.volume ?? 0);
	const avgVolume = calculateAverageVolume(history);

	if (currentVolume === 0 || avgVolume === 0) return undefined;

	const ratio = currentVolume / avgVolume;

	if (ratio > 2) {
		return 'on very high volume';
	}
	if (ratio > 1.5) {
		return 'on above-average volume';
	}
	if (ratio < 0.5) {
		return 'on thin volume';
	}

	return undefined;
}

/**
 * Generate the complete price story
 */
export function generatePriceStory(context: PriceContext): PriceStory {
	const { current, history } = context;
	const currentPrice = current.ltp ?? current.close ?? 0;
	const { high, low } = get52WeekRange(history);
	const rangePosition = calculateRangePosition(currentPrice, high, low);

	return {
		positionSentence: generatePositionSentence(currentPrice, high, low),
		trendSentence: generateTrendSentence(history),
		volumeContext: generateVolumeContext(current, history),
		rangePosition
	};
}
