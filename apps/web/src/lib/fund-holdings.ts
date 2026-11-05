import type { Fund, Holding } from '#lib/types/fund.js';

export interface FundHolding {
	fundSymbol: string;
	fundName: string;
	units: number | undefined;
	value: number;
	percentOfFund: number;
}

function escapeRegExp(string: string): string {
	return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function normalizeName(name: string): string {
	return name
		.toLowerCase()
		.replace(/hydro\s+power/g, 'hydropower')
		.replace(/\b(limited|ltd|pvt|private|company|co|inc|corporation|development)\b/g, '')
		.replace(/[^\w\s]/g, '')
		.replace(/\s+/g, ' ')
		.trim();
}

export function findCompanyInFunds(funds: Fund[], companyName: string): FundHolding[] {
	const resultsMap = new Map<string, FundHolding>();
	const normalizedTarget = normalizeName(companyName);

	// Create regex for word boundary matching
	const targetRegex = new RegExp(`\\b${escapeRegExp(normalizedTarget)}\\b`, 'i');

	for (const fund of funds) {
		// Search all holding categories
		for (const [category, holdings] of Object.entries(fund.holdings)) {
			// Skip non-equity categories to avoid confusion (e.g. matching "Sanima Bank Debenture" when looking for "Sanima Bank")
			if (
				category === 'fixed_deposits' ||
				category === 'government_bonds' ||
				category === 'corporate_debentures'
			) {
				continue;
			}

			if (!Array.isArray(holdings)) continue;

			for (const holding of holdings as Holding[]) {
				const holdingName = holding.name;
				const normalizedHolding = normalizeName(holdingName);

				// Create reverse regex (if holding is shorter than target)
				const holdingRegex = new RegExp(`\\b${escapeRegExp(normalizedHolding)}\\b`, 'i');

				// Check both directions
				if (targetRegex.test(normalizedHolding) || holdingRegex.test(normalizedTarget)) {
					// Safety check: Avoid very short matches
					if (normalizedTarget.length < 4 || normalizedHolding.length < 4) {
						continue;
					}

					const existing = resultsMap.get(fund.symbol);
					if (existing) {
						// Aggregate if already found (e.g. in another category)
						existing.units = (existing.units || 0) + (holding.units || 0);
						existing.value += holding.value;
						existing.percentOfFund = (existing.value / fund.net_assets) * 100;
					} else {
						resultsMap.set(fund.symbol, {
							fundSymbol: fund.symbol,
							fundName: fund.fund_name,
							units: holding.units,
							value: holding.value,
							percentOfFund: (holding.value / fund.net_assets) * 100
						});
					}
					break; // Found in this category, move to next category for this fund
				}
			}
		}
	}

	// Sort by value descending
	return Array.from(resultsMap.values()).sort((a, b) => b.value - a.value);
}
