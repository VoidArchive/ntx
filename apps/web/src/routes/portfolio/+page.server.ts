import { Code, ConnectError } from '@connectrpc/connect';
import { fail, redirect } from '@sveltejs/kit';
import { apiBaseUrl, createApiClient } from '#lib/api/client.js';
import navData from '#lib/data/nav.json';
import { HoldingStatus } from '#lib/gen/ntx/v1/portfolio_pb.js';
import { OWNER_COOKIE, forgetOwner, rememberOwner } from '#lib/owner.js';
import { forward, holdingActions, num, str } from '#lib/owner-actions.js';
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
	...holdingActions,
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
				// Absent on an add: the API picks the sector's default.
				drawdownGate: form.has('drawdown_gate_known')
					? form.get('drawdown_gate') === 'on'
					: undefined,
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
	// The owner confirms or edits each proposed carry; one number per symbol.
	executed: async ({ request, cookies }) => {
		const form = await request.formData();
		const carries: Record<string, number> = {};
		// A cleared field is left alone; only a number the owner typed is written.
		for (const [key, value] of form.entries()) {
			if (key.startsWith('carry:') && String(value).trim() !== '')
				carries[key.slice('carry:'.length)] = Number(value);
		}
		return forward(cookies, (api) => api.markMonthExecuted({ carries }));
	},
	import: async ({ request, cookies }) => {
		const file = (await request.formData()).get('file');
		if (!(file instanceof File) || !file.size)
			return fail(400, { message: 'Choose the file to import.' });
		const json = new Uint8Array(await file.arrayBuffer());
		return forward(cookies, (api) => api.importPersonal({ json }));
	},
	rules: async ({ request, cookies }) => {
		const form = await request.formData();
		return forward(cookies, (api) =>
			api.updateRules({
				rules: {
					softCapPercent: num(form, 'soft_cap_percent'),
					hardCapPercent: num(form, 'hard_cap_percent'),
					sectorCapPercent: num(form, 'sector_cap_percent'),
					maxPositions: Math.trunc(num(form, 'max_positions')),
					minEntryRs: num(form, 'min_entry_rs'),
					oneHydroPerBasin: form.get('one_hydro_per_basin') === 'on',
					candidateMinRating: str(form, 'candidate_min_rating'),
					candidateDividendRecord: form.get('candidate_dividend_record') === 'on',
					monthlyBudgetRs: num(form, 'monthly_budget_rs'),
					multiplierBelowMedian: num(form, 'multiplier_below_median'),
					multiplierBottomQuarter: num(form, 'multiplier_bottom_quarter')
				}
			})
		);
	}
};
