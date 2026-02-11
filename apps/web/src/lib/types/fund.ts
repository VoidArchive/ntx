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
	mutual_funds: 'Mutual Funds',
	investment: 'Investment',
	trading: 'Trading',
	fixed_deposits: 'Fixed Deposits',
	government_bonds: 'Govt. Bonds',
	corporate_debentures: 'Debentures'
};

// Warm-to-cool gradient palette for visual harmony
export const SECTOR_COLORS: Record<keyof Holdings, string> = {
	fixed_deposits: '#f97316', // orange (primary - largest)
	commercial_banks: '#fb923c', // orange-400
	life_insurance: '#fbbf24', // amber-400
	hydropower: '#a3e635', // lime-400
	others: '#4ade80', // green-400
	manufacturing: '#2dd4bf', // teal-400
	development_banks: '#22d3ee', // cyan-400
	microfinance: '#38bdf8', // sky-400
	non_life_insurance: '#60a5fa', // blue-400
	finance_companies: '#818cf8', // indigo-400
	mutual_funds: '#a78bfa', // violet-400
	investment: '#c084fc', // purple-400
	trading: '#e879f9', // fuchsia-400
	government_bonds: '#f0abfc', // fuchsia-300
	corporate_debentures: '#d8b4fe', // purple-300
	hotels: '#f9a8d4' // pink-300
};
