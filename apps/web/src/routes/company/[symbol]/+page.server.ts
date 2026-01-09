import { company, price } from '$lib/api/client';
import { Code, ConnectError } from '@connectrpc/connect';
import { error } from '@sveltejs/kit';

export const load = async ({ params }) => {
	try {
		const [companyRes, fundamentalsRes, priceRes, priceHistoryRes] = await Promise.all([
			company.getCompany({ symbol: params.symbol }),
			company.getFundamentals({ symbol: params.symbol }),
			price.getPrice({ symbol: params.symbol }),
			price.getPriceHistory({ symbol: params.symbol, days: 365 })
		]);

		// Fetch sector stats (non-blocking - we can still show page without it)
		let sectorStats = undefined;
		if (companyRes.company?.sector) {
			try {
				const sectorRes = await company.getSectorStats({ sector: companyRes.company.sector });
				sectorStats = sectorRes.stats;
			} catch {
				// Sector stats are optional, continue without them
			}
		}

		return {
			company: companyRes.company,
			fundamentals: fundamentalsRes.latest,
			fundamentalsHistory: fundamentalsRes.history,
			price: priceRes.price,
			priceHistory: priceHistoryRes.prices,
			sectorStats
		};
	} catch (err) {
		if (err instanceof ConnectError && err.code === Code.NotFound) {
			throw error(404, 'equity not found');
		}
		throw error(500, err instanceof Error ? err.message : 'Unknown error');
	}
};
