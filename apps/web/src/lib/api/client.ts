import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { CompanyService } from '$lib/gen/ntx/v1/company_pb';
import { PriceService } from '$lib/gen/ntx/v1/price_pb';
import { browser } from '$app/environment';

// Determine base URL: Internal (SSR) vs Public (Browser)
// 'process' is available in Node.js environment (adapter-node)
let baseUrl = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';
if (!browser && typeof process !== 'undefined' && process.env.INTERNAL_API_URL) {
	baseUrl = process.env.INTERNAL_API_URL;
}

const transport = createConnectTransport({
	baseUrl
});

export const company = createClient(CompanyService, transport);
export const price = createClient(PriceService, transport);
