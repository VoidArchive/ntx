import { createApiClient } from '#lib/api/client.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
	const { apiUrl } = await parent();
	const { market } = createApiClient(apiUrl);
	// The market line is a header, not the page: an API hiccup should hide it, not 500 the home page.
	try {
		const res = await market.getMarketDay({});
		return { marketDay: res.marketDay };
	} catch {
		return { marketDay: undefined };
	}
};
