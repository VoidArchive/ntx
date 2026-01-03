import type { PageServerLoad } from './$types';
import { company } from '$lib/api/client';
import { Sector } from '@ntx/api/ntx/v1/common_pb';

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
		const response = await company.listCompanies({ sector, query });
		return {
			companies: response.companies,
			sector,
			query
		};
	} catch (err) {
		console.error('Failed to load companies:', err);
		return {
			companies: [],
			sector,
			query
		};
	}
};
