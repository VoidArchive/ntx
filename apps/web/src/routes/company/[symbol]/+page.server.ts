import { apiBaseUrl, createApiClient } from '#lib/api/client.js';
import navData from '#lib/data/nav.json';
import { Code, ConnectError } from '@connectrpc/connect';
import { error } from '@sveltejs/kit';
import type { Fund } from '#lib/types/fund.js';
import { findCompanyInFunds, type FundHolding } from '#lib/fund-holdings.js';
import { OWNER_COOKIE } from '#lib/owner.js';
import { holdingActions } from '#lib/owner-actions.js';
import type { GetHoldingStoryResponse } from '#lib/gen/ntx/v1/portfolio_pb.js';

// The owner's thesis for this company, or nothing: not held, no token, or a wrong one.
async function holdingStory(apiUrl: string, token: string | undefined, symbol: string) {
	if (!token) return undefined;
	try {
		return await createApiClient(apiUrl, token).portfolio.getHoldingStory({
			symbol,
			fundNavs: Object.fromEntries((navData as Fund[]).map((f) => [f.symbol, f.nav_per_unit]))
		});
	} catch (e) {
		if (e instanceof ConnectError && [Code.NotFound, Code.Unauthenticated].includes(e.code))
			return undefined;
		throw e;
	}
}

export const actions = holdingActions;

export const load = async ({ params, cookies }) => {
	const apiUrl = apiBaseUrl();
	const { company, price, story, hydro } = createApiClient(apiUrl);
	try {
		const [companyRes, fundamentalsRes, priceHistoryRes, plantsRes, thesis] = await Promise.all([
			company.getCompany({ symbol: params.symbol }),
			company.getFundamentals({ symbol: params.symbol }),
			price.getPriceHistory({ symbol: params.symbol, days: 365 }),
			// Plants are a section, not the page: a hydro outage must not 500 every company.
			hydro.getCompanyPlants({ symbol: params.symbol }).catch(() => ({ plants: [], opening: '' })),
			holdingStory(apiUrl, cookies.get(OWNER_COOKIE), params.symbol) as Promise<
				GetHoldingStoryResponse | undefined
			>
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
			thesis,
			navMonths: [...new Set((navData as Fund[]).map((f) => f.report_date_nepali))].join(', ')
		};
	} catch (err) {
		if (err instanceof ConnectError && err.code === Code.NotFound) {
			throw error(404, 'equity not found');
		}
		throw error(500, err instanceof Error ? err.message : 'Unknown error');
	}
};
