import { company, price } from '$lib/api/client';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	const [companyRes, priceRes] = await Promise.all([
		company.listCompanies({ limit: 500 }),
		price.listLatestPrices({})
	]);

	return {
		companies: companyRes.companies,
		prices: priceRes.prices
	};
};
