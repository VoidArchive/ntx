import type { Fund } from '$lib/types/fund';

export async function load({ fetch }) {
	const response = await fetch('/data/nav_detailed.json');
	const funds: Fund[] = await response.json();

	// Sort by net assets descending
	funds.sort((a, b) => b.net_assets - a.net_assets);

	return { funds };
}
