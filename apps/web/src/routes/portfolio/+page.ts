import type { PageLoad } from './$types';
import { createApiClient } from '$lib/api/client';

export const load: PageLoad = async ({ parent }) => {
	const { apiUrl } = await parent();
	const api = createApiClient(apiUrl);
	
	// Fetch companies for the stock symbol autocomplete
	try {
		const response = await api.company.listCompanies({ limit: 500 });
		return {
			companies: response.companies
		};
	} catch {
		return {
			companies: []
		};
	}
};
