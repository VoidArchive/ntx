import { apiBaseUrl, createApiClient } from '#lib/api/client.js';
import { OWNER_COOKIE } from '#lib/owner.js';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies }) => {
	const apiUrl = apiBaseUrl();
	const { company } = createApiClient(apiUrl);
	const { companies } = await company.listCompanies({ limit: 500 });

	// Every page carries this, so the search gets symbol and name and nothing else.
	// `owner` only decides whether the sidebar shows Portfolio; the API checks the token itself.
	return {
		search: companies.map(({ symbol, name }) => ({ symbol, name })),
		apiUrl,
		owner: cookies.get(OWNER_COOKIE) !== undefined
	};
};
