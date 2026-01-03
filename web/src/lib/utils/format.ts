// Nepali Rupee formatting
export function formatPrice(value: number): string {
	return `Rs. ${value.toLocaleString('en-NP', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

export function formatPriceCompact(value: number): string {
	return value.toLocaleString('en-NP', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

export function formatChange(value: number): string {
	const sign = value >= 0 ? '+' : '';
	return `${sign}${value.toFixed(2)}%`;
}

export function formatChangeWithValue(change: number, percentChange: number): string {
	const sign = change >= 0 ? '+' : '';
	return `${sign}${change.toFixed(2)} (${sign}${percentChange.toFixed(2)}%)`;
}

export function formatVolume(value: bigint | number): string {
	const num = typeof value === 'bigint' ? Number(value) : value;
	if (num >= 1_000_000_000) return `${(num / 1_000_000_000).toFixed(2)}B`;
	if (num >= 1_000_000) return `${(num / 1_000_000).toFixed(2)}M`;
	if (num >= 1_000) return `${(num / 1_000).toFixed(1)}K`;
	return num.toLocaleString();
}

export function formatMarketCap(value: number): string {
	if (value >= 1_000_000_000_000) return `Rs. ${(value / 1_000_000_000_000).toFixed(2)}T`;
	if (value >= 1_000_000_000) return `Rs. ${(value / 1_000_000_000).toFixed(2)}B`;
	if (value >= 1_000_000) return `Rs. ${(value / 1_000_000).toFixed(2)}M`;
	return `Rs. ${value.toLocaleString()}`;
}

export function formatNumber(value: number, decimals = 2): string {
	return value.toLocaleString('en-NP', {
		minimumFractionDigits: decimals,
		maximumFractionDigits: decimals
	});
}

export function formatPercent(value: number): string {
	return `${value.toFixed(2)}%`;
}

export function formatDate(date: Date | string): string {
	const d = typeof date === 'string' ? new Date(date) : date;
	return d.toLocaleDateString('en-NP', {
		year: 'numeric',
		month: 'short',
		day: 'numeric'
	});
}

export function formatDateShort(date: Date | string): string {
	const d = typeof date === 'string' ? new Date(date) : date;
	return d.toLocaleDateString('en-NP', {
		month: 'short',
		day: 'numeric'
	});
}
