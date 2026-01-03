import type { PageServerLoad } from './$types';
import { market, screener } from '$lib/api/client';

export const load: PageServerLoad = async () => {
	try {
		const [statusRes, indicesRes, sectorsRes, gainersRes, losersRes] = await Promise.allSettled([
			market.getStatus({}),
			market.listIndices({}),
			market.listSectors({}),
			screener.listTopGainers({ limit: 5 }),
			screener.listTopLosers({ limit: 5 })
		]);

		return {
			status: statusRes.status === 'fulfilled' ? statusRes.value.status : null,
			indices: indicesRes.status === 'fulfilled' ? indicesRes.value.indices : [],
			sectors: sectorsRes.status === 'fulfilled' ? sectorsRes.value.sectors : [],
			gainers: gainersRes.status === 'fulfilled' ? gainersRes.value.stocks : [],
			losers: losersRes.status === 'fulfilled' ? losersRes.value.stocks : []
		};
	} catch (err) {
		console.error('Failed to load market data:', err);
		return {
			status: null,
			indices: [],
			sectors: [],
			gainers: [],
			losers: []
		};
	}
};
