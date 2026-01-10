import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { CompanyService } from '$lib/gen/ntx/v1/company_pb';
import { PriceService } from '$lib/gen/ntx/v1/price_pb';

export function createApiClient(baseUrl: string) {
	const transport = createConnectTransport({
		baseUrl,
		fetch: (input, init) => fetch(input, { ...init, redirect: 'follow' })
	});

	return {
		company: createClient(CompanyService, transport),
		price: createClient(PriceService, transport)
	};
}
