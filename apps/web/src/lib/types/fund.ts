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
	total_assets: number;
	total_liabilities: number;
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
	fixed_deposits: 'Fixed Deposits',
	government_bonds: 'Govt. Bonds',
	corporate_debentures: 'Debentures'
};

export const SECTOR_COLORS: Record<keyof Holdings, string> = {
	commercial_banks: '#3b82f6', // blue
	development_banks: '#6366f1', // indigo
	finance_companies: '#8b5cf6', // violet
	life_insurance: '#ec4899', // pink
	non_life_insurance: '#f43f5e', // rose
	hydropower: '#14b8a6', // teal
	hotels: '#f97316', // orange
	manufacturing: '#eab308', // yellow
	microfinance: '#22c55e', // green
	others: '#64748b', // slate
	fixed_deposits: '#0ea5e9', // sky
	government_bonds: '#06b6d4', // cyan
	corporate_debentures: '#a855f7' // purple
};
