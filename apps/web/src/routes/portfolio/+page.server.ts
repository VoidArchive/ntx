import { Code, ConnectError } from '@connectrpc/connect';
import { fail, redirect } from '@sveltejs/kit';
import { apiBaseUrl, createApiClient } from '#lib/api/client.js';
import navData from '#lib/data/nav.json';
import { HoldingStatus, TransactionKind } from '#lib/gen/ntx/v1/portfolio_pb.js';
import { OWNER_COOKIE, forgetOwner, rememberOwner } from '#lib/owner.js';
import type { Fund } from '#lib/types/fund.js';
import type { Actions, PageServerLoad } from './$types';

const funds = navData as Fund[];

export const load: PageServerLoad = async ({ parent, cookies }) => {
	const token = cookies.get(OWNER_COOKIE);
	if (!token) return { needsToken: true as const };

	const { apiUrl } = await parent();
	const { portfolio } = createApiClient(apiUrl, token);
	try {
		const res = await portfolio.getPortfolio({
			fundNavs: Object.fromEntries(funds.map((f) => [f.symbol, f.nav_per_unit]))
		});
		return {
			needsToken: false as const,
			portfolio: res,
			navDates: Object.fromEntries(funds.map((f) => [f.symbol, f.report_date_english]))
		};
	} catch (e) {
		// A wrong token reads the same as none: back to the token field.
		if (e instanceof ConnectError && e.code === Code.Unauthenticated) {
			forgetOwner(cookies);
			return { needsToken: true as const };
		}
		throw e;
	}
};

const str = (form: FormData, key: string) => String(form.get(key) ?? '').trim();
const num = (form: FormData, key: string) => Number(form.get(key) ?? 0) || 0;

// Each action forwards one RPC; the API validates and the message comes back verbatim.
async function forward(
	cookies: import('@sveltejs/kit').Cookies,
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

async function isListed(symbol: string) {
	if (funds.some((f) => f.symbol === symbol)) return true;
	try {
		await createApiClient(apiBaseUrl()).company.getCompany({ symbol });
		return true;
	} catch (e) {
		if (e instanceof ConnectError && e.code === Code.NotFound) return false;
		throw e;
	}
}

export const actions: Actions = {
	token: async ({ request, cookies }) => {
		const token = str(await request.formData(), 'token');
		if (!token) return fail(400, { message: 'The token is empty.' });
		rememberOwner(cookies, token);
		redirect(303, '/portfolio');
	},
	forget: async ({ cookies }) => {
		forgetOwner(cookies);
		redirect(303, '/');
	},
	holding: async ({ request, cookies }) => {
		const form = await request.formData();
		const symbol = str(form, 'symbol').toUpperCase();
		// A symbol the API has never heard of would otherwise be stored as a priceless fund.
		if (!(await isListed(symbol)))
			return fail(400, { message: `${symbol} is not a listed company or fund.` });
		return forward(cookies, (api) =>
			api.upsertHolding({
				symbol,
				status: num(form, 'status') as HoldingStatus,
				targetShares: BigInt(Math.trunc(num(form, 'target_shares'))),
				allocationPercent: num(form, 'allocation_percent'),
				carryRs: num(form, 'carry_rs'),
				drawdownGate: form.get('drawdown_gate') === 'on',
				thesis: str(form, 'thesis'),
				frozenReason: str(form, 'frozen_reason'),
				reviewDate: str(form, 'review_date')
			})
		);
	},
	delete: async ({ request, cookies }) => {
		const form = await request.formData();
		return forward(cookies, (api) => api.deleteHolding({ symbol: str(form, 'symbol') }));
	},
	transaction: async ({ request, cookies }) => {
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
	note: async ({ request, cookies }) => {
		const form = await request.formData();
		return forward(cookies, (api) =>
			api.addNote({ symbol: str(form, 'symbol'), day: str(form, 'day'), text: str(form, 'text') })
		);
	}
};
