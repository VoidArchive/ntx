import type { PageServerLoad } from './$types';
import { screener } from '$lib/api/client';
import { Sector } from '@ntx/api/ntx/v1/common_pb';
import { SortBy, SortOrder } from '@ntx/api/ntx/v1/screener_pb';

export const load: PageServerLoad = async ({ url }) => {
	const sectorParam = url.searchParams.get('sector');
	const query = url.searchParams.get('q') ?? '';

	let sector = Sector.UNSPECIFIED;
	if (sectorParam) {
		const sectorNum = parseInt(sectorParam, 10);
		if (!isNaN(sectorNum) && sectorNum >= 0 && sectorNum <= 13) {
			sector = sectorNum as Sector;
		}
	}

	try {
		// Use screener API to get companies with their prices
		const response = await screener.screen({
			sector,
			sortBy: SortBy.SYMBOL,
			sortOrder: SortOrder.ASC,
			limit: 500,
			offset: 0
		});

		// Filter by query if provided
		let results = response.results;
		if (query) {
			const q = query.toLowerCase();
			results = results.filter(
				(r) =>
					r.company?.symbol.toLowerCase().includes(q) ||
					r.company?.name.toLowerCase().includes(q)
			);
		}

		return {
			results,
			sector,
			query,
			total: results.length
		};
	} catch (err) {
		console.error('Failed to load companies:', err);
		return {
			results: [],
			sector,
			query,
			total: 0
		};
	}
};
