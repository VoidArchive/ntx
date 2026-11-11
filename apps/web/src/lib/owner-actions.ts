import { ConnectError } from '@connectrpc/connect';
import { fail, type Cookies } from '@sveltejs/kit';
import { apiBaseUrl, createApiClient } from '#lib/api/client.js';
import { GateDirection, GateKind, TransactionKind } from '#lib/gen/ntx/v1/portfolio_pb.js';
import { OWNER_COOKIE } from '#lib/owner.js';

export const str = (form: FormData, key: string) => String(form.get(key) ?? '').trim();
export const num = (form: FormData, key: string) => Number(form.get(key) ?? 0) || 0;
// An empty field is not recorded, never a zero (ADR 0002).
export const opt = (form: FormData, key: string) =>
	str(form, key) === '' ? undefined : num(form, key);

// Each action forwards one RPC; the API validates and the message comes back verbatim.
export async function forward(
	cookies: Cookies,
	call: (api: ReturnType<typeof createApiClient>['portfolio']) => Promise<unknown>
) {
	const token = cookies.get(OWNER_COOKIE);
	if (!token) return fail(401, { message: 'Enter the owner token first.' });
	try {
		await call(createApiClient(apiBaseUrl(), token).portfolio);
		return { ok: true };
	} catch (e) {
		if (e instanceof ConnectError) return fail(400, { message: e.rawMessage || e.code.toString() });
		throw e;
	}
}

// The rows a holding carries on both /portfolio and its company page.
export const holdingActions = {
	transaction: async ({ request, cookies }: { request: Request; cookies: Cookies }) => {
		const form = await request.formData();
		return forward(cookies, (api) =>
			api.addTransaction({
				symbol: str(form, 'symbol'),
				day: str(form, 'day'),
				kind: num(form, 'kind') as TransactionKind,
				quantity: BigInt(Math.trunc(num(form, 'quantity'))),
				price: num(form, 'price')
			})
		);
	},
	note: async ({ request, cookies }: { request: Request; cookies: Cookies }) => {
		const form = await request.formData();
		return forward(cookies, (api) =>
			api.addNote({ symbol: str(form, 'symbol'), day: str(form, 'day'), text: str(form, 'text') })
		);
	},
	gate: async ({ request, cookies }: { request: Request; cookies: Cookies }) => {
		const form = await request.formData();
		return forward(cookies, (api) =>
			api.upsertGate({
				id: BigInt(Math.trunc(num(form, 'id'))),
				symbol: str(form, 'symbol'),
				kind: num(form, 'kind') as GateKind,
				metric: str(form, 'metric'),
				direction: num(form, 'direction') as GateDirection,
				threshold: num(form, 'threshold'),
				periods: Math.trunc(num(form, 'periods')),
				text: str(form, 'text'),
				checkedOn: str(form, 'checked_on')
			})
		);
	},
	deleteGate: async ({ request, cookies }: { request: Request; cookies: Cookies }) => {
		const form = await request.formData();
		return forward(cookies, (api) => api.deleteGate({ id: BigInt(Math.trunc(num(form, 'id'))) }));
	},
	bankMetric: async ({ request, cookies }: { request: Request; cookies: Cookies }) => {
		const form = await request.formData();
		return forward(cookies, (api) =>
			api.upsertBankMetric({
				metric: {
					symbol: str(form, 'symbol'),
					fiscalYear: str(form, 'fiscal_year'),
					quarter: str(form, 'quarter'),
					nplPercent: opt(form, 'nplPercent'),
					carPercent: opt(form, 'carPercent'),
					loanGrowthPercent: opt(form, 'loanGrowthPercent'),
					distributableProfit: opt(form, 'distributableProfit'),
					publishedDate: str(form, 'published_date'),
					sourceUrl: str(form, 'source_url')
				}
			})
		);
	}
};
