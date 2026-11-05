import { apiBaseUrl, createApiClient } from '#lib/api/client.js';
import navData from '#lib/data/nav.json';
import { Code, ConnectError } from '@connectrpc/connect';
import { error } from '@sveltejs/kit';
import type { Fund } from '#lib/types/fund.js';
import { findCompanyInFunds, type FundHolding } from '#lib/fund-holdings.js';

export const load = async ({ params }) => {
	const apiUrl = apiBaseUrl();
	const { company, price, story, hydro } = createApiClient(apiUrl);
	try {
		const [companyRes, fundamentalsRes, priceHistoryRes, plantsRes] = await Promise.all([
			company.getCompany({ symbol: params.symbol }),
			company.getFundamentals({ symbol: params.symbol }),
			price.getPriceHistory({ symbol: params.symbol, days: 365 }),
			hydro.getCompanyPlants({ symbol: params.symbol })
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
			plants: plantsRes.plants,
			plantsOpening: plantsRes.opening,
			navMonths: [...new Set((navData as Fund[]).map((f) => f.report_date_nepali))].join(', ')
		};
	} catch (err) {
		if (err instanceof ConnectError && err.code === Code.NotFound) {
			throw error(404, 'equity not found');
		}
		throw error(500, err instanceof Error ? err.message : 'Unknown error');
	}
};
