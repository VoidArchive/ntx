import type { Fund } from '#lib/types/fund.js';

// NEPSE open-end funds issue units at Rs 10
const PAR = 10;

function fmtLarge(value: number): string {
	if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`;
	if (value >= 10_000_000) return `${(value / 10_000_000).toFixed(2)} Cr`;
	if (value >= 100_000) return `${(value / 100_000).toFixed(2)} L`;
	return value.toLocaleString('en-NP');
}

// Sign and glyph carry the direction so colour is never the only signal
function parStanding(navPerUnit: number): string {
	const pct = ((navPerUnit - PAR) / PAR) * 100;
	if (pct > 0) return `▲ +${pct.toFixed(2)}% above the Rs ${PAR} par`;
	if (pct < 0) return `▼ ${pct.toFixed(2)}% below the Rs ${PAR} par`;
	return `at the Rs ${PAR} par`;
}

export function fundOpeningSentence(fund: Fund): string {
	return `As of ${fund.report_date_english} (${fund.report_date_nepali}), ${fund.symbol}'s NAV was Rs ${fund.nav_per_unit.toFixed(2)} per unit, ${parStanding(fund.nav_per_unit)}, on net assets of Rs ${fmtLarge(fund.net_assets)}.`;
}

export function fundProvenance(fund: Fund): string {
	return `Figures come from the fund's NAV report for ${fund.report_date_nepali} (${fund.report_date_english}), a static snapshot rather than live data.`;
}
