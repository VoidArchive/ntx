import { error, json } from '@sveltejs/kit';
import { apiBaseUrl, createApiClient } from '#lib/api/client.js';
import type { RequestHandler } from './$types';

// The browser never sees API_URL, so an expanding row fetches its seven days through here.
export const GET: RequestHandler = async ({ params }) => {
	if (!/^\d{1,9}$/.test(params.id)) throw error(400, 'plant id must be a number');
	const { hydro } = createApiClient(apiBaseUrl());
	const res = await hydro.getPlantReadings({ plantId: BigInt(params.id) });
	return json(
		res.readings.map((r) => ({
			day: r.day,
			rainMm: r.rainMm,
			dischargeM3s: r.dischargeM3s,
			rainSource: r.rainSource,
			dischargeSource: r.dischargeSource,
			state: r.state
		}))
	);
};
