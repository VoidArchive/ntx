import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { MarketService } from '@ntx/api/ntx/v1/market_pb';
import { CompanyService } from '@ntx/api/ntx/v1/company_pb';
import { PriceService } from '@ntx/api/ntx/v1/price_pb';
import { ScreenerService } from '@ntx/api/ntx/v1/screener_pb';

const transport = createConnectTransport({
	baseUrl: import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
});

export const market = createClient(MarketService, transport);
export const company = createClient(CompanyService, transport);
export const price = createClient(PriceService, transport);
export const screener = createClient(ScreenerService, transport);
