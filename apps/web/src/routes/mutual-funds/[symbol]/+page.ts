import type { Fund } from '$lib/types/fund';
import { error } from '@sveltejs/kit';
import navData from '$lib/data/nav_2026_01.json';

export async function load({ params }) {
	const funds: Fund[] = navData as Fund[];

	const fund = funds.find((f) => f.symbol === params.symbol.toUpperCase());

	if (!fund) {
		throw error(404, `Fund ${params.symbol} not found`);
	}

	return { fund, funds };
}
