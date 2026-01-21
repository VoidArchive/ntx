import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { CompanyService } from '$lib/gen/ntx/v1/company_pb';
import { PriceService } from '$lib/gen/ntx/v1/price_pb';
import { AuthService } from '$lib/gen/ntx/v1/auth_pb';
import { PortfolioService } from '$lib/gen/ntx/v1/portfolio_pb';

export function createApiClient(baseUrl: string, getToken?: () => string | null) {
	const transport = createConnectTransport({
		baseUrl,
		fetch: (input, init) => {
			const token = getToken?.();
			const headers = new Headers(init?.headers);
			if (token) {
				headers.set('Authorization', `Bearer ${token}`);
			}
			return fetch(input, { ...init, headers, redirect: 'follow' });
		}
	});

	return {
		company: createClient(CompanyService, transport),
		price: createClient(PriceService, transport),
		auth: createClient(AuthService, transport),
		portfolio: createClient(PortfolioService, transport)
	};
}

