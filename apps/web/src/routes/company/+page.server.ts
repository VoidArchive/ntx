import { createApiClient } from '#lib/api/client.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
	const { apiUrl } = await parent();
	const { company } = createApiClient(apiUrl);
	const { rows } = await company.listScreener({});
	return { rows };
};
