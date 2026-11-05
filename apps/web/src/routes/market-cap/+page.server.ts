import { listCompaniesWithLatestPrices } from '#lib/api/client.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
	const { apiUrl } = await parent();
	return listCompaniesWithLatestPrices(apiUrl);
};
