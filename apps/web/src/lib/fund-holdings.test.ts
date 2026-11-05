import { describe, expect, it } from 'vitest';
import { findCompanyInFunds } from './fund-holdings.js';
import type { Fund } from './types/fund.js';

// One fund, net assets 1000, so a holding's value is its percent times ten.
const fund: Fund = {
	symbol: 'NIBLSF',
	fund_name: 'NIBL Sahabhagita Fund',
	fund_manager: 'NIBL Ace Capital',
	report_date_nepali: '2082-09',
	report_date_english: '2026-01',
	nav_per_unit: 10,
	total_units: 100,
	net_assets: 1000,
	holdings_disclosure: 'full',
	holdings: {
		commercial_banks: [
			{ name: 'Nabil Bank Limited', units: 10, value: 100 },
			{ name: 'Sanima Bank Ltd', units: 5, value: 50 },
			{ name: 'NLG', units: 1, value: 10 }
		],
		corporate_debentures: [{ name: 'Sanima Bank Debenture 2085', value: 300 }]
	}
};

describe('findCompanyInFunds', () => {
	it('matches a holding by exact name', () => {
		expect(findCompanyInFunds([fund], 'Nabil Bank Limited')).toEqual([
			{
				fundSymbol: 'NIBLSF',
				fundName: 'NIBL Sahabhagita Fund',
				units: 10,
				value: 100,
				percentOfFund: 10
			}
		]);
	});

	it('matches after stripping Limited and Ltd', () => {
		expect(findCompanyInFunds([fund], 'Sanima Bank Limited').map((h) => h.value)).toEqual([50]);
	});

	it('does not match a holding whose stripped name is too short', () => {
		expect(findCompanyInFunds([fund], 'NLG Insurance Company Ltd')).toEqual([]);
	});

	it('skips debentures so a bank does not inherit its bond', () => {
		const [hit] = findCompanyInFunds([fund], 'Sanima Bank');
		expect(hit.value).toBe(50);
		expect(hit.percentOfFund).toBe(5);
	});

	it('returns nothing for a company no fund holds', () => {
		expect(findCompanyInFunds([fund], 'Upper Tamakoshi Hydropower')).toEqual([]);
	});
});
