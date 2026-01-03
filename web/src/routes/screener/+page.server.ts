import type { PageServerLoad } from './$types';
import { screener } from '$lib/api/client';
import { Sector } from '@ntx/api/ntx/v1/common_pb';
import { SortBy, SortOrder } from '@ntx/api/ntx/v1/screener_pb';

export const load: PageServerLoad = async ({ url }) => {
	// Parse URL parameters
	const sectorParam = url.searchParams.get('sector');
	const minPrice = url.searchParams.get('minPrice');
	const maxPrice = url.searchParams.get('maxPrice');
	const minPe = url.searchParams.get('minPe');
	const maxPe = url.searchParams.get('maxPe');
	const minPb = url.searchParams.get('minPb');
	const maxPb = url.searchParams.get('maxPb');
	const minChange = url.searchParams.get('minChange');
	const maxChange = url.searchParams.get('maxChange');
	const sortByParam = url.searchParams.get('sort');
	const sortOrderParam = url.searchParams.get('order');
	const near52wHigh = url.searchParams.get('near52wHigh') === 'true';
	const near52wLow = url.searchParams.get('near52wLow') === 'true';
	const limitParam = url.searchParams.get('limit');
	const offsetParam = url.searchParams.get('offset');

	// Build request
	const request: Parameters<typeof screener.screen>[0] = {
		sector: Sector.UNSPECIFIED,
		sortBy: SortBy.UNSPECIFIED,
		sortOrder: SortOrder.DESC,
		limit: 50,
		offset: 0
	};

	if (sectorParam) {
		const sectorNum = parseInt(sectorParam, 10);
		if (!isNaN(sectorNum)) request.sector = sectorNum as Sector;
	}

	if (minPrice) request.minPrice = parseFloat(minPrice);
	if (maxPrice) request.maxPrice = parseFloat(maxPrice);
	if (minPe) request.minPe = parseFloat(minPe);
	if (maxPe) request.maxPe = parseFloat(maxPe);
	if (minPb) request.minPb = parseFloat(minPb);
	if (maxPb) request.maxPb = parseFloat(maxPb);
	if (minChange) request.minChange = parseFloat(minChange);
	if (maxChange) request.maxChange = parseFloat(maxChange);
	if (near52wHigh) request.near52wHigh = true;
	if (near52wLow) request.near52wLow = true;
	if (limitParam) request.limit = parseInt(limitParam, 10);
	if (offsetParam) request.offset = parseInt(offsetParam, 10);

	// Parse sort
	const sortByMap: Record<string, SortBy> = {
		symbol: SortBy.SYMBOL,
		price: SortBy.PRICE,
		change: SortBy.CHANGE,
		volume: SortBy.VOLUME,
		turnover: SortBy.TURNOVER,
		marketCap: SortBy.MARKET_CAP,
		pe: SortBy.PE
	};
	if (sortByParam && sortByMap[sortByParam]) {
		request.sortBy = sortByMap[sortByParam];
	}
	if (sortOrderParam === 'asc') {
		request.sortOrder = SortOrder.ASC;
	}

	try {
		const response = await screener.screen(request);
		return {
			results: response.results,
			total: response.total,
			filters: {
				sector: request.sector,
				minPrice: request.minPrice,
				maxPrice: request.maxPrice,
				minPe: request.minPe,
				maxPe: request.maxPe,
				minPb: request.minPb,
				maxPb: request.maxPb,
				minChange: request.minChange,
				maxChange: request.maxChange,
				near52wHigh: request.near52wHigh,
				near52wLow: request.near52wLow,
				sortBy: sortByParam ?? '',
				sortOrder: sortOrderParam ?? 'desc',
				limit: request.limit ?? 50,
				offset: request.offset ?? 0
			}
		};
	} catch (err) {
		console.error('Failed to run screener:', err);
		return {
			results: [],
			total: 0,
			filters: {
				sector: Sector.UNSPECIFIED,
				sortBy: '',
				sortOrder: 'desc',
				limit: 50,
				offset: 0
			}
		};
	}
};
