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

	const targetRegex = new RegExp(`\\b${escapeRegExp(normalizedTarget)}\\b`, 'i');

	for (const fund of funds) {
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

				const holdingRegex = new RegExp(`\\b${escapeRegExp(normalizedHolding)}\\b`, 'i');

				if (targetRegex.test(normalizedHolding) || holdingRegex.test(normalizedTarget)) {
					if (normalizedTarget.length < 4 || normalizedHolding.length < 4) {
						continue;
					}

					// A fund can hold the same company in two categories; the row aggregates them.
					const existing = resultsMap.get(fund.symbol) ?? {
						fundSymbol: fund.symbol,
						fundName: fund.fund_name,
						units: undefined,
						value: 0,
						percentOfFund: 0
					};
					existing.units = (existing.units ?? 0) + (holding.units ?? 0);
					existing.value += holding.value;
					existing.percentOfFund = (existing.value / fund.net_assets) * 100;
					resultsMap.set(fund.symbol, existing);
					break; // Found in this category, move to next category for this fund
				}
			}
		}
	}

	return Array.from(resultsMap.values()).sort((a, b) => b.value - a.value);
}
