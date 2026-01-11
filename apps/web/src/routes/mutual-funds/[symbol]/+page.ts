import type { Fund } from '$lib/types/fund';
import { error } from '@sveltejs/kit';

export async function load({ params, fetch }) {
	const response = await fetch('/data/nav_detailed.json');
	const funds: Fund[] = await response.json();

	const fund = funds.find((f) => f.symbol === params.symbol.toUpperCase());

	if (!fund) {
		throw error(404, `Fund ${params.symbol} not found`);
	}

	return { fund, funds };
}
