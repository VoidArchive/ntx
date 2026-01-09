/**
 * Verdict Story Generator
 *
 * Combines price, earnings, and valuation signals into an overall narrative.
 */

import type { PriceStory, EarningsStory, ValuationStory, VerdictStory } from './types';

type Signal = 'opportunity' | 'caution' | 'neutral';

interface VerdictContext {
  price: PriceStory;
  earnings: EarningsStory;
  valuation: ValuationStory;
  companyName: string;
}

/**
 * Determine price signal
 */
function getPriceSignal(price: PriceStory): Signal {
  // Near lows = potential opportunity
  if (price.rangePosition < 0.3) return 'opportunity';
  // Near highs = caution
  if (price.rangePosition > 0.8) return 'caution';
  return 'neutral';
}

/**
 * Determine earnings signal
 */
function getEarningsSignal(earnings: EarningsStory): Signal {
  const growth = earnings.epsGrowth;
  if (growth === undefined) return 'neutral';

  if (growth > 10 || earnings.growthStreak >= 2) return 'opportunity';
  if (growth < -10) return 'caution';
  return 'neutral';
}

/**
 * Determine valuation signal
 */
function getValuationSignal(valuation: ValuationStory): Signal {
  const percentile = valuation.sectorPercentile;
  if (percentile === undefined) return 'neutral';

  if (percentile < 40) return 'opportunity'; // Cheap
  if (percentile > 70) return 'caution'; // Expensive
  return 'neutral';
}

/**
 * Combine signals into overall verdict
 */
function combineSignals(
  priceSignal: Signal,
  earningsSignal: Signal,
  valuationSignal: Signal
): Signal {
  const signals = [priceSignal, earningsSignal, valuationSignal];
  const opportunities = signals.filter((s) => s === 'opportunity').length;
  const cautions = signals.filter((s) => s === 'caution').length;

  // Strong opportunity: 2+ opportunity signals, no caution
  if (opportunities >= 2 && cautions === 0) return 'opportunity';

  // Strong caution: 2+ caution signals
  if (cautions >= 2) return 'caution';

  // Opportunity if more positive than negative
  if (opportunities > cautions) return 'opportunity';

  // Caution if more negative than positive
  if (cautions > opportunities) return 'caution';

  return 'neutral';
}

/**
 * Generate summary sentence based on signals
 */
function generateSummary(
  context: VerdictContext,
  overallSignal: Signal
): string {
  const { companyName, price, earnings, valuation } = context;

  // Build the summary sentence
  const parts: string[] = [];

  // Price context
  if (price.rangePosition < 0.3) {
    parts.push(`${companyName} is trading near its 52-week lows`);
  } else if (price.rangePosition > 0.7) {
    parts.push(`${companyName} is trading near its 52-week highs`);
  } else {
    parts.push(`${companyName} is in the middle of its 52-week range`);
  }

  // Earnings context
  if (earnings.growthStreak >= 2) {
    parts.push(`with a track record of consistent earnings growth`);
  } else if (earnings.epsGrowth !== undefined && earnings.epsGrowth > 0) {
    parts.push(`with growing earnings`);
  } else if (earnings.epsGrowth !== undefined && earnings.epsGrowth < 0) {
    parts.push(`despite declining earnings`);
  }

  // Valuation context
  if (valuation.sectorPercentile !== undefined) {
    if (valuation.sectorPercentile < 40) {
      parts.push(`The valuation suggests the market may be underpricing its fundamentals.`);
    } else if (valuation.sectorPercentile > 70) {
      parts.push(`The premium valuation reflects high expectations.`);
    }
  }

  let summary = parts.slice(0, 2).join(' ') + '. ';
  if (parts[2]) {
    summary += parts[2];
  }

  // Add disclaimer
  return summary;
}

/**
 * Generate the complete verdict story
 */
export function generateVerdictStory(context: VerdictContext): VerdictStory {
  const priceSignal = getPriceSignal(context.price);
  const earningsSignal = getEarningsSignal(context.earnings);
  const valuationSignal = getValuationSignal(context.valuation);

  const signal = combineSignals(priceSignal, earningsSignal, valuationSignal);

  return {
    title: 'The Story So Far',
    summary: generateSummary(context, signal),
    signal
  };
}
