import type { PageLoad } from './$types';
import { createApiClient } from '$lib/api/client';

const API_URL = import.meta.env.DEV ? 'http://localhost:8080' : 'https://ntx-api.anishshrestha.com';

export const load: PageLoad = async () => {
	const api = createApiClient(API_URL);
	
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
