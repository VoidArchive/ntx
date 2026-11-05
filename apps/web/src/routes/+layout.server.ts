import { apiBaseUrl, createApiClient } from '#lib/api/client.js';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async () => {
	const apiUrl = apiBaseUrl();
	const { company } = createApiClient(apiUrl);
	const { companies } = await company.listCompanies({ limit: 500 });

	// Every page carries this, so the search gets symbol and name and nothing else.
	return {
		search: companies.map(({ symbol, name }) => ({ symbol, name })),
		apiUrl
	};
};
