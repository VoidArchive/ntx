import type { Fund } from '$lib/types/fund';
import navData from '$lib/data/nav_2026_01.json';

export async function load() {
	const funds: Fund[] = navData as Fund[];

	// Sort by net assets descending
	funds.sort((a, b) => b.net_assets - a.net_assets);

	return { funds };
}
