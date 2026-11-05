import { createApiClient, listMarket } from '#lib/api/client.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
	const { apiUrl } = await parent();
	const { market, sector } = createApiClient(apiUrl);
	// Each section hides on an API hiccup rather than 500ing the home page.
	const [marketDay, overview, { companies, prices }] = await Promise.all([
		market
			.getMarketDay({})
			.then((res) => res.marketDay)
			.catch(() => undefined),
		sector.getSectorOverview({}).catch(() => undefined),
		listMarket(apiUrl).catch(() => ({ companies: [], prices: [] }))
	]);
	return { marketDay, overview, companies, prices };
};
