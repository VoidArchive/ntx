<script lang="ts">
	import { formatLarge } from '#lib/utils.js';
	import type { Fund } from '#lib/types/fund.js';

	interface Props {
		fund: Fund;
	}

	let { fund }: Props = $props();

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

	function optional(label: string, value: number | undefined): StatRow[] {
		if (value === undefined) return [];
		return [{ label, value: formatLarge(value) }];
	}

	// Balance sheet totals are only present in reports that publish them
	let stats = $derived.by((): StatRow[] => [
		{ label: 'Net Assets', value: formatLarge(fund.net_assets) },
		...optional('Total Assets', fund.total_assets),
		...optional('Total Liabilities', fund.total_liabilities),
		{ label: 'Total Units', value: formatLarge(fund.total_units) },
		{ label: 'Holdings', value: holdingsLabel },
		...(fund.holdings_disclosure === 'full'
			? [{ label: 'Sectors', value: fmtNumber(sectorsCount) }]
			: [])
	]);

	let holdingsLabel = $derived(
		fund.holdings_disclosure === 'full'
			? fmtNumber(holdingsCount)
			: fund.holdings_disclosure === 'top10'
				? 'Top 10 only'
				: 'Not published'
	);
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
