export interface Holding {
	name: string;
	units?: number;
	value: number;
}

export interface Holdings {
	commercial_banks?: Holding[];
	development_banks?: Holding[];
	finance_companies?: Holding[];
	life_insurance?: Holding[];
	non_life_insurance?: Holding[];
	hydropower?: Holding[];
	hotels?: Holding[];
	manufacturing?: Holding[];
	microfinance?: Holding[];
	others?: Holding[];
	mutual_funds?: Holding[];
	investment?: Holding[];
	trading?: Holding[];
	fixed_deposits?: Holding[];
	government_bonds?: Holding[];
	corporate_debentures?: Holding[];
}

export interface Fund {
	symbol: string;
	fund_name: string;
	fund_manager: string;
	report_date_nepali: string;
	report_date_english: string;
	nav_per_unit: number;
	total_units: number;
	net_assets: number;
	// Funds publish all holdings, only their 10 largest, or none; the rest is one "Other listed shares" row
	holdings_disclosure: 'full' | 'top10' | 'none';
	total_assets?: number;
	total_liabilities?: number;
	holdings: Holdings;
}

export const SECTOR_LABELS: Record<keyof Holdings, string> = {
	commercial_banks: 'Commercial Banks',
	development_banks: 'Development Banks',
	finance_companies: 'Finance',
	life_insurance: 'Life Insurance',
	non_life_insurance: 'Non-Life Insurance',
	hydropower: 'Hydropower',
	hotels: 'Hotels',
	manufacturing: 'Manufacturing',
	microfinance: 'Microfinance',
	others: 'Others',
	mutual_funds: 'Mutual Funds',
	investment: 'Investment',
	trading: 'Trading',
	fixed_deposits: 'Fixed Deposits',
	government_bonds: 'Govt. Bonds',
	corporate_debentures: 'Debentures'
};

// Warm-to-cool gradient palette for visual harmony
