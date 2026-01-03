import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { company, price, screener } from '$lib/api/client';
import { Sector } from '@ntx/api/ntx/v1/common_pb';
import { SortBy, SortOrder } from '@ntx/api/ntx/v1/screener_pb';

export const load: PageServerLoad = async ({ params }) => {
	const { symbol } = params;

	try {
		const [companyRes, priceRes, fundamentalsRes, reportsRes, candlesRes] = await Promise.allSettled([
			company.getCompany({ symbol }),
			price.getPrice({ symbol }),
			company.getFundamentals({ symbol }),
			company.listReports({ symbol, limit: 20 }),
			price.listCandles({ symbol }) // Get price history
		]);

		const companyData =
			companyRes.status === 'fulfilled' ? companyRes.value.company : null;

		if (!companyData) {
			throw error(404, `Company ${symbol} not found`);
		}

		// Get sector peers
		type ScreenResult = Awaited<ReturnType<typeof screener.screen>>['results'][0];
		let sectorPeers: ScreenResult[] = [];
		if (companyData.sector !== Sector.UNSPECIFIED) {
			try {
				const screenerRes = await screener.screen({
					sector: companyData.sector,
					sortBy: SortBy.MARKET_CAP,
					sortOrder: SortOrder.DESC,
					limit: 10,
					offset: 0
				});
				// Filter out the current company from peers
				sectorPeers = screenerRes.results.filter(r => r.company?.symbol !== symbol);
			} catch {
				// Silently fail for peers
			}
		}

		return {
			company: companyData,
			price: priceRes.status === 'fulfilled' ? priceRes.value.price : null,
			fundamentals:
				fundamentalsRes.status === 'fulfilled' ? fundamentalsRes.value.fundamentals : null,
			reports: reportsRes.status === 'fulfilled' ? reportsRes.value.reports : [],
			candles: candlesRes.status === 'fulfilled' ? candlesRes.value.candles : [],
			sectorPeers
		};
	} catch (err) {
		if (err && typeof err === 'object' && 'status' in err) {
			throw err;
		}
		console.error('Failed to load company data:', err);
		throw error(500, 'Failed to load company data');
	}
};
