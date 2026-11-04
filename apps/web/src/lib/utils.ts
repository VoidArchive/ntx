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

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, 'child'> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, 'children'> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };
