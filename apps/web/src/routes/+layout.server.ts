import { createApiClient } from '$lib/api/client';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ platform }) => {
	const apiUrl = import.meta.env.DEV
		? 'http://localhost:8080'
		: (platform?.env?.API_URL ?? 'http://localhost:8080');
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
