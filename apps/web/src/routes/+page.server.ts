import { createApiClient } from '$lib/api/client';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ platform }) => {
	const apiUrl = platform?.env?.API_URL ?? 'http://localhost:8080';
	const { company, price } = createApiClient(apiUrl);

	const [companyRes, priceRes] = await Promise.all([
		company.listCompanies({ limit: 500 }),
		price.listLatestPrices({})
	]);

	return {
		companies: companyRes.companies,
		prices: priceRes.prices
	};
};
