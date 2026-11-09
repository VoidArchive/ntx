import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { Sector } from '#lib/gen/ntx/v1/common_pb.js';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export const sectorNames: Record<number, string> = {
	[Sector.COMMERCIAL_BANK]: 'Commercial Banking',
	[Sector.DEVELOPMENT_BANK]: 'Development Banking',
	[Sector.FINANCE]: 'Finance',
	[Sector.MICROFINANCE]: 'Microfinance',
	[Sector.LIFE_INSURANCE]: 'Life Insurance',
	[Sector.NON_LIFE_INSURANCE]: 'Non-Life Insurance',
	[Sector.HYDROPOWER]: 'Hydropower',
	[Sector.MANUFACTURING]: 'Manufacturing',
	[Sector.HOTEL]: 'Hotels & Tourism',
	[Sector.TRADING]: 'Trading',
	[Sector.INVESTMENT]: 'Investment',
	[Sector.MUTUAL_FUND]: 'Mutual Fund',
	[Sector.OTHERS]: 'Others'
};

// Large rupee amounts and counts in arba, crore and lakh.
export function formatLarge(value: number, digits = 2): string {
	if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(digits)} Arba`;
	if (value >= 10_000_000) return `${(value / 10_000_000).toFixed(digits)} Cr`;
	if (value >= 100_000) return `${(value / 100_000).toFixed(digits)} L`;
	return value.toLocaleString('en-NP');
}

// Position on a strip, 0 to 100, with the median pinned at 50: the half below the median fills
// the left half and the half above fills the right, so one outlier cannot crush the rest against
// an end and "left of centre" always reads as below the median.
export function at(range: { low: number; median: number; high: number }, value: number): number {
	if (value <= range.median) {
		const span = range.median - range.low;
		return span <= 0 ? 50 : Math.max(0, (50 * (value - range.low)) / span);
	}
	const span = range.high - range.median;
	return span <= 0 ? 50 : Math.min(100, 50 + (50 * (value - range.median)) / span);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, 'child'> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, 'children'> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };
