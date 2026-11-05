import { createApiClient } from '#lib/api/client.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
	const { apiUrl } = await parent();
	const { hydro } = createApiClient(apiUrl);
	const res = await hydro.listPlants({});
	return {
		plants: res.plants,
		basins: res.basins,
		rules: res.rules,
		lastRun: res.lastRun,
		opening: res.opening
	};
};
