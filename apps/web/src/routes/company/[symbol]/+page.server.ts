import { company } from '$lib/api/client';
import { Code, ConnectError } from '@connectrpc/connect';
import { error } from '@sveltejs/kit';

export const load = async ({ params }) => {
	try {
		const res = await company.getCompany({ symbol: params.symbol });
		return { company: res.company };
	} catch (err) {
		if (err instanceof ConnectError && err.code === Code.NotFound) {
			throw error(404, 'equity not found');
		}
		throw error(500, err instanceof Error ? err.message : 'Unknown error');
	}
};
