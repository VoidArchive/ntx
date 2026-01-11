<script lang="ts">
	import type { Fund } from '$lib/types/fund';

	interface Props {
		fund: Fund;
	}

	let { fund }: Props = $props();

	function fmtLarge(value: number): string {
		if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`;
		if (value >= 10_000_000) return `${(value / 10_000_000).toFixed(2)} Cr`;
		if (value >= 100_000) return `${(value / 100_000).toFixed(2)} L`;
		return value.toLocaleString('en-NP');
	}

	function fmtNumber(value: number): string {
		return value.toLocaleString('en-NP');
	}

	// Count holdings
	let holdingsCount = $derived.by(() => {
		let count = 0;
		for (const items of Object.values(fund.holdings)) {
			if (Array.isArray(items)) {
				count += items.length;
			}
		}
		return count;
	});

	// Count sectors
	let sectorsCount = $derived.by(() => {
		let count = 0;
		for (const items of Object.values(fund.holdings)) {
			if (Array.isArray(items) && items.length > 0) {
				count++;
			}
		}
		return count;
	});

	interface StatRow {
		label: string;
		value: string;
	}

	let stats = $derived.by((): StatRow[] => [
		{ label: 'Net Assets', value: fmtLarge(fund.net_assets) },
		{ label: 'Total Assets', value: fmtLarge(fund.total_assets) },
		{ label: 'Total Liabilities', value: fmtLarge(fund.total_liabilities) },
		{ label: 'Total Units', value: fmtLarge(fund.total_units) },
		{ label: 'NAV/Unit', value: `Rs. ${fund.nav_per_unit.toFixed(2)}` },
		{ label: 'Holdings', value: fmtNumber(holdingsCount) },
		{ label: 'Sectors', value: fmtNumber(sectorsCount) }
	]);
</script>

<div class="text-sm">
	{#each stats as stat, i (stat.label)}
		<div
			class="flex justify-between py-2.5 {i < stats.length - 1
				? 'border-b border-dotted border-border'
				: ''}"
		>
			<span class="text-muted-foreground">{stat.label}</span>
			<span class="font-medium tabular-nums">{stat.value}</span>
		</div>
	{/each}
</div>
