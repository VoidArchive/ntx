/**
 * Story Engine - Main Entry Point
 *
 * Generates complete company narratives from raw data.
 */

export type {
	StoryData,
	PriceStory,
	EarningsStory,
	ValuationStory,
	VerdictStory,
	CompanyStory
} from './types';

export { generatePriceStory } from './price';
export { generateEarningsStory } from './earnings';
export { generateValuationStory } from './valuation';
export { generateVerdictStory } from './verdict';

import type { StoryData, CompanyStory } from './types';
import { generatePriceStory } from './price';
import { generateEarningsStory } from './earnings';
import { generateValuationStory } from './valuation';
import { generateVerdictStory } from './verdict';

/**
 * Generate a complete company story from all available data
 */
export function generateStory(data: StoryData): CompanyStory {
	const priceStory = generatePriceStory({
		current: data.price,
		history: data.priceHistory
	});

	const earningsStory = generateEarningsStory({
		current: data.fundamentals,
		history: data.fundamentalsHistory
	});

	const valuationStory = generateValuationStory({
		current: data.fundamentals,
		sectorStats: data.sectorStats
	});

	const verdictStory = generateVerdictStory({
		price: priceStory,
		earnings: earningsStory,
		valuation: valuationStory,
		companyName: data.company.name
	});

	return {
		price: priceStory,
		earnings: earningsStory,
		valuation: valuationStory,
		verdict: verdictStory
	};
}

/**
 * Generate a mini-story (one-liner) for company cards
 */
export function generateMiniStory(data: StoryData): string {
	const priceStory = generatePriceStory({
		current: data.price,
		history: data.priceHistory
	});

	const earningsStory = generateEarningsStory({
		current: data.fundamentals,
		history: data.fundamentalsHistory
	});

	// Combine key insights into one sentence
	const pricePart =
		priceStory.rangePosition < 0.3
			? 'Near 52W low'
			: priceStory.rangePosition > 0.7
				? 'Near 52W high'
				: '';

	const earningsPart =
		earningsStory.growthStreak >= 2
			? 'consistent growth'
			: earningsStory.epsGrowth !== undefined && earningsStory.epsGrowth > 0
				? 'earnings up'
				: earningsStory.epsGrowth !== undefined && earningsStory.epsGrowth < 0
					? 'earnings down'
					: '';

	if (pricePart && earningsPart) {
		return `${pricePart}, ${earningsPart}`;
	}

	return pricePart || earningsPart || 'View details →';
}
