/**
 * Story Engine - Types
 *
 * Common types used across all story generation modules.
 */

import type { Company, Price, Fundamental } from '$lib/gen/ntx/v1/common_pb';
import type { SectorStats } from '$lib/gen/ntx/v1/company_pb';

/**
 * All data needed to generate a company's story
 */
export interface StoryData {
	company: Company;
	price: Price;
	priceHistory: Price[];
	fundamentals: Fundamental;
	fundamentalsHistory: Fundamental[];
	sectorStats?: SectorStats;
}

/**
 * Generated price-related insights
 */
export interface PriceStory {
	/** E.g., "Trading 25% below 52-week high" */
	positionSentence: string;
	/** E.g., "Down 10% over the past month" */
	trendSentence: string;
	/** E.g., "on above-average volume" */
	volumeContext?: string;
	/** Position in 52W range: 0 = at low, 1 = at high */
	rangePosition: number;
}

/**
 * Generated earnings-related insights
 */
export interface EarningsStory {
	/** E.g., "Earnings are growing faster than the stock price" */
	headline: string;
	/** E.g., "EPS increased 15% this year while the price fell 25%" */
	detail: string;
	/** E.g., "Four consecutive years of profit growth" */
	trendSentence: string;
	/** YoY EPS growth rate */
	epsGrowth?: number;
	/** Number of consecutive growth years */
	growthStreak: number;
}

/**
 * Generated valuation-related insights
 */
export interface ValuationStory {
	/** E.g., "Cheaper than 70% of commercial banks" */
	headline: string;
	/** E.g., "P/E Ratio: 17.5" */
	peContext: string;
	/** Percentile rank within sector (0-100) */
	sectorPercentile?: number;
}

/**
 * Overall company verdict
 */
export interface VerdictStory {
	/** E.g., "The Story So Far" */
	title: string;
	/** 2-3 sentence summary combining all signals */
	summary: string;
	/** Signal classification */
	signal: 'opportunity' | 'caution' | 'neutral';
}

/**
 * Complete generated story for a company
 */
export interface CompanyStory {
	price: PriceStory;
	earnings: EarningsStory;
	valuation: ValuationStory;
	verdict: VerdictStory;
}
