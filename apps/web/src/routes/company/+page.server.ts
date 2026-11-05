import { createApiClient } from '#lib/api/client.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
	const { apiUrl } = await parent();
	const { company } = createApiClient(apiUrl);
	const { companies } = await company.listCompanies({ limit: 500 });
	return { companies };
};
