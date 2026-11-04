import { apiBaseUrl, createApiClient } from '#lib/api/client.js';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async () => {
	const apiUrl = apiBaseUrl();
	const { company, price } = createApiClient(apiUrl);

	const [companyRes, priceRes] = await Promise.all([
		company.listCompanies({ limit: 500 }),
		price.listLatestPrices({})
	]);

	return {
		companies: companyRes.companies,
		prices: priceRes.prices,
		apiUrl
	};
};
