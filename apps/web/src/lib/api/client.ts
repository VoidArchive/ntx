import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { CompanyService } from '#lib/gen/ntx/v1/company_pb.js';
import { MarketService } from '#lib/gen/ntx/v1/market_pb.js';
import { PriceService } from '#lib/gen/ntx/v1/price_pb.js';
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
		market: createClient(MarketService, transport)
	};
}
