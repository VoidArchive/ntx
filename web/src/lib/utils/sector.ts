import { Sector } from '@ntx/api/ntx/v1/common_pb';

export const sectorNames: Record<Sector, string> = {
	[Sector.UNSPECIFIED]: 'All Sectors',
	[Sector.COMMERCIAL_BANK]: 'Commercial Banks',
	[Sector.DEVELOPMENT_BANK]: 'Development Banks',
	[Sector.FINANCE]: 'Finance',
	[Sector.MICROFINANCE]: 'Microfinance',
	[Sector.LIFE_INSURANCE]: 'Life Insurance',
	[Sector.NON_LIFE_INSURANCE]: 'Non-Life Insurance',
	[Sector.HYDROPOWER]: 'Hydropower',
	[Sector.MANUFACTURING]: 'Manufacturing',
	[Sector.HOTEL]: 'Hotels',
	[Sector.TRADING]: 'Trading',
	[Sector.INVESTMENT]: 'Investment',
	[Sector.MUTUAL_FUND]: 'Mutual Funds',
	[Sector.OTHERS]: 'Others'
};

export const sectorColors: Record<Sector, string> = {
	[Sector.UNSPECIFIED]: 'bg-muted',
	[Sector.COMMERCIAL_BANK]: 'bg-blue-500/10 text-blue-700 dark:text-blue-300',
	[Sector.DEVELOPMENT_BANK]: 'bg-indigo-500/10 text-indigo-700 dark:text-indigo-300',
	[Sector.FINANCE]: 'bg-violet-500/10 text-violet-700 dark:text-violet-300',
	[Sector.MICROFINANCE]: 'bg-purple-500/10 text-purple-700 dark:text-purple-300',
	[Sector.LIFE_INSURANCE]: 'bg-pink-500/10 text-pink-700 dark:text-pink-300',
	[Sector.NON_LIFE_INSURANCE]: 'bg-rose-500/10 text-rose-700 dark:text-rose-300',
	[Sector.HYDROPOWER]: 'bg-cyan-500/10 text-cyan-700 dark:text-cyan-300',
	[Sector.MANUFACTURING]: 'bg-orange-500/10 text-orange-700 dark:text-orange-300',
	[Sector.HOTEL]: 'bg-amber-500/10 text-amber-700 dark:text-amber-300',
	[Sector.TRADING]: 'bg-lime-500/10 text-lime-700 dark:text-lime-300',
	[Sector.INVESTMENT]: 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300',
	[Sector.MUTUAL_FUND]: 'bg-teal-500/10 text-teal-700 dark:text-teal-300',
	[Sector.OTHERS]: 'bg-slate-500/10 text-slate-700 dark:text-slate-300'
};

export function getSectorName(sector: Sector): string {
	return sectorNames[sector] ?? 'Unknown';
}

export function getSectorColor(sector: Sector): string {
	return sectorColors[sector] ?? sectorColors[Sector.UNSPECIFIED];
}

// Get all sectors excluding UNSPECIFIED for filters/dropdowns
export function getAllSectors(): Sector[] {
	return [
		Sector.COMMERCIAL_BANK,
		Sector.DEVELOPMENT_BANK,
		Sector.FINANCE,
		Sector.MICROFINANCE,
		Sector.LIFE_INSURANCE,
		Sector.NON_LIFE_INSURANCE,
		Sector.HYDROPOWER,
		Sector.MANUFACTURING,
		Sector.HOTEL,
		Sector.TRADING,
		Sector.INVESTMENT,
		Sector.MUTUAL_FUND,
		Sector.OTHERS
	];
}
