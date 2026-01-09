import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { CompanyService } from '$lib/gen/ntx/v1/company_pb';
import { PriceService } from '$lib/gen/ntx/v1/price_pb';

const transport = createConnectTransport({
	baseUrl: import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
});

export const company = createClient(CompanyService, transport);
export const price = createClient(PriceService, transport);
