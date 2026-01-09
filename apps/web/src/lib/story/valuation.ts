/**
 * Valuation Story Generator
 *
 * Generates narrative insights about a company's valuation relative to sector.
 */

import { Sector, type Fundamental } from '$lib/gen/ntx/v1/common_pb';
import type { SectorStats } from '$lib/gen/ntx/v1/company_pb';
import type { ValuationStory } from './types';

interface ValuationContext {
  current: Fundamental;
  sectorStats?: SectorStats;
}

/**
 * Map sector enum to display name
 */
const sectorDisplayNames: Record<number, string> = {
  [Sector.COMMERCIAL_BANK]: 'commercial banks',
  [Sector.DEVELOPMENT_BANK]: 'development banks',
  [Sector.FINANCE]: 'finance companies',
  [Sector.MICROFINANCE]: 'microfinance institutions',
  [Sector.LIFE_INSURANCE]: 'life insurers',
  [Sector.NON_LIFE_INSURANCE]: 'non-life insurers',
  [Sector.HYDROPOWER]: 'hydropower companies',
  [Sector.MANUFACTURING]: 'manufacturing firms',
  [Sector.HOTEL]: 'hotels',
  [Sector.TRADING]: 'trading companies',
  [Sector.INVESTMENT]: 'investment companies',
  [Sector.MUTUAL_FUND]: 'mutual funds',
  [Sector.OTHERS]: 'peers'
};

/**
 * Get sector display name
 */
function getSectorName(sectorStats?: SectorStats): string {
  if (!sectorStats) return 'peers';
  return sectorDisplayNames[sectorStats.sector] ?? 'peers';
}


/**
 * Calculate percentile (how company compares to sector average)
 * Returns 0-100 where higher means more expensive (for P/E)
 */
function calculatePePercentile(companyPe: number, sectorAvgPe: number): number {
  // Simple heuristic: if company P/E < sector avg, it's cheaper
  // This would ideally use full sector distribution, but we only have average
  const ratio = companyPe / sectorAvgPe;

  // Convert ratio to approximate percentile
  // 0.5x avg = ~10th percentile, 1.0x = ~50th, 2.0x = ~90th
  if (ratio <= 0.5) return 10;
  if (ratio <= 0.75) return 30;
  if (ratio <= 1.0) return 50;
  if (ratio <= 1.25) return 70;
  if (ratio <= 1.5) return 85;
  return 95;
}

/**
 * Generate headline based on P/E comparison
 */
function generateHeadline(
  companyPe: number | undefined,
  sectorStats?: SectorStats
): string {
  if (companyPe === undefined) {
    return 'Valuation data not available';
  }

  const sectorAvgPe = sectorStats?.avgPeRatio;
  const sectorName = getSectorName(sectorStats);

  if (sectorAvgPe === undefined) {
    if (companyPe < 10) return 'Trading at a low P/E multiple';
    if (companyPe < 20) return 'Trading at a moderate P/E multiple';
    return 'Trading at a high P/E multiple';
  }

  const percentile = calculatePePercentile(companyPe, sectorAvgPe);

  if (percentile < 30) {
    return `Cheaper than most ${sectorName}`;
  }

  if (percentile < 50) {
    return `Below-average valuation for ${sectorName}`;
  }

  if (percentile < 70) {
    return `Fairly valued among ${sectorName}`;
  }

  if (percentile < 85) {
    return `Above-average valuation for ${sectorName}`;
  }

  return `Premium valuation among ${sectorName}`;
}

/**
 * Generate P/E context sentence
 */
function generatePeContext(
  companyPe: number | undefined,
  sectorStats?: SectorStats
): string {
  if (companyPe === undefined) {
    return 'P/E ratio not available.';
  }

  const sectorAvgPe = sectorStats?.avgPeRatio;

  if (sectorAvgPe === undefined) {
    return `P/E Ratio: ${companyPe.toFixed(1)}`;
  }

  const diff = ((companyPe - sectorAvgPe) / sectorAvgPe) * 100;
  const comparison = diff > 0 ? 'above' : 'below';

  return `P/E Ratio: ${companyPe.toFixed(1)} (${Math.abs(diff).toFixed(0)}% ${comparison} sector average of ${sectorAvgPe.toFixed(1)})`;
}

/**
 * Generate the complete valuation story
 */
export function generateValuationStory(context: ValuationContext): ValuationStory {
  const { current, sectorStats } = context;
  const companyPe = current.peRatio;
  const sectorAvgPe = sectorStats?.avgPeRatio;

  const sectorPercentile =
    companyPe !== undefined && sectorAvgPe !== undefined
      ? calculatePePercentile(companyPe, sectorAvgPe)
      : undefined;

  return {
    headline: generateHeadline(companyPe, sectorStats),
    peContext: generatePeContext(companyPe, sectorStats),
    sectorPercentile
  };
}
