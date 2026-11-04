import { createApiClient } from '#lib/api/client.js';
import navData from '#lib/data/nav.json';
import { Code, ConnectError } from '@connectrpc/connect';
import { error } from '@sveltejs/kit';
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

function findCompanyInFunds(funds: Fund[], companyName: string): FundHolding[] {
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

export const load = async ({ params, platform }) => {
	// In dev mode, always use localhost. In prod, use platform env.
	const apiUrl = import.meta.env.DEV
		? 'http://localhost:8080'
		: (platform?.env?.API_URL ?? 'http://localhost:8080');
	const { company, price, story } = createApiClient(apiUrl);
	try {
		const [companyRes, fundamentalsRes, priceHistoryRes] = await Promise.all([
			company.getCompany({ symbol: params.symbol }),
			company.getFundamentals({ symbol: params.symbol }),
			price.getPriceHistory({ symbol: params.symbol, days: 365 })
		]);

		// Fetch mutual fund holdings for this company
		let fundHoldings: FundHolding[] = [];
		if (companyRes.company?.name) {
			try {
				const funds = navData as Fund[];
				fundHoldings = findCompanyInFunds(funds, companyRes.company.name);
			} catch {
				// Fund holdings are optional
			}
		}

		// The story engine formats the fund sentence; it only gets the summary.
		const largest = fundHoldings[0];
		const storyRes = await story.getCompanyStory({
			symbol: params.symbol,
			funds: largest && {
				count: fundHoldings.length,
				totalValue: fundHoldings.reduce((sum, h) => sum + h.value, 0),
				largestName: largest.fundName,
				largestValue: largest.value,
				asOf:
					(navData as Fund[]).find((f) => f.symbol === largest.fundSymbol)?.report_date_nepali ?? ''
			}
		});

		return {
			company: companyRes.company,
			story: storyRes,
			fundamentalsHistory: fundamentalsRes.history,
			priceHistory: priceHistoryRes.prices,
			fundHoldings,
			navMonths: [...new Set((navData as Fund[]).map((f) => f.report_date_nepali))].join(', ')
		};
	} catch (err) {
		if (err instanceof ConnectError && err.code === Code.NotFound) {
			throw error(404, 'equity not found');
		}
		throw error(500, err instanceof Error ? err.message : 'Unknown error');
	}
};
