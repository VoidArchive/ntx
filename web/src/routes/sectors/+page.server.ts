import type { PageServerLoad } from './$types';
import { market } from '$lib/api/client';

export const load: PageServerLoad = async () => {
	try {
		const response = await market.listSectors({});
		return {
			sectors: response.sectors
		};
	} catch (err) {
		console.error('Failed to load sectors:', err);
		return {
			sectors: []
		};
	}
};
