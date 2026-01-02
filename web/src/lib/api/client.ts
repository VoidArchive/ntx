import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { MarketService } from '@ntx/api/ntx/v1/market_pb';
import { AnalyzerService } from '@ntx/api/ntx/v1/analyzer_pb';

const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
});

export const market = createClient(MarketService, transport);
export const analyzer = createClient(AnalyzerService, transport);
