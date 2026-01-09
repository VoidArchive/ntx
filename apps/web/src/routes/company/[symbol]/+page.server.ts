import { company, price } from '$lib/api/client';
import { Code, ConnectError } from '@connectrpc/connect';
import { error } from '@sveltejs/kit';

export const load = async ({ params }) => {
	try {
		const [companyRes, fundamentalsRes, priceRes] = await Promise.all([
			company.getCompany({ symbol: params.symbol }),
			company.getFundamentals({ symbol: params.symbol }),
			price.getPrice({ symbol: params.symbol })
		]);
		return {
			company: companyRes.company,
			fundamentals: fundamentalsRes.latest,
			fundamentalsHistory: fundamentalsRes.history,
			price: priceRes.price
		};
	} catch (err) {
		if (err instanceof ConnectError && err.code === Code.NotFound) {
			throw error(404, 'equity not found');
		}
		throw error(500, err instanceof Error ? err.message : 'Unknown error');
	}
};
