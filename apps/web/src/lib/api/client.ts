import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { API_URL } from '$app/env/private';
import { CompanyService } from '#lib/gen/ntx/v1/company_pb.js';
import { HydroService } from '#lib/gen/ntx/v1/hydro_pb.js';
import { MarketService } from '#lib/gen/ntx/v1/market_pb.js';
import { PriceService } from '#lib/gen/ntx/v1/price_pb.js';
import { SectorService } from '#lib/gen/ntx/v1/sector_pb.js';
import { StoryService } from '#lib/gen/ntx/v1/story_pb.js';

export function createApiClient(baseUrl: string) {
	const transport = createConnectTransport({
		baseUrl,
		fetch: (input, init) => fetch(input, { ...init, redirect: 'follow' })
	});

	return {
		company: createClient(CompanyService, transport),
		price: createClient(PriceService, transport),
		story: createClient(StoryService, transport),
		market: createClient(MarketService, transport),
		sector: createClient(SectorService, transport),
		hydro: createClient(HydroService, transport)
	};
}

// The market tables (home, market cap, most traded) rank every company by its latest price.
export async function listCompaniesWithLatestPrices(apiUrl: string) {
	const { company, price } = createApiClient(apiUrl);
	const [companyRes, priceRes] = await Promise.all([
		company.listCompanies({ limit: 500 }),
		price.listLatestPrices({})
	]);
	return { companies: companyRes.companies, prices: priceRes.prices };
}

export const apiBaseUrl = () => {
	if (API_URL) return API_URL;
	if (import.meta.env.DEV) return 'http://localhost:8080';
	// Never fall back to localhost in production: that is how every page 500ed on 2026-11-04.
	throw new Error('API_URL is not set');
};
