<script lang="ts">
	import { formatNumber, formatPercent, formatMarketCap } from '$lib/utils/format';
	import type { Fundamentals } from '@ntx/api/ntx/v1/common_pb';

	let { fundamentals }: { fundamentals: Fundamentals } = $props();

	const metrics = $derived([
		{ label: 'P/E Ratio', value: fundamentals.pe, format: (v: number) => formatNumber(v, 2) },
		{ label: 'P/B Ratio', value: fundamentals.pb, format: (v: number) => formatNumber(v, 2) },
		{ label: 'EPS', value: fundamentals.eps, format: (v: number) => `Rs. ${formatNumber(v, 2)}` },
		{
			label: 'Book Value',
			value: fundamentals.bookValue,
			format: (v: number) => `Rs. ${formatNumber(v, 2)}`
		},
		{ label: 'Market Cap', value: fundamentals.marketCap, format: formatMarketCap },
		{
			label: 'Dividend Yield',
			value: fundamentals.dividendYield,
			format: (v: number) => formatPercent(v)
		},
		{ label: 'ROE', value: fundamentals.roe, format: (v: number) => formatPercent(v) },
		{
			label: 'Shares Outstanding',
			value: Number(fundamentals.sharesOutstanding),
			format: (v: number) => v.toLocaleString()
		}
	]);
</script>

<div class="grid grid-cols-2 gap-4 md:grid-cols-4">
	{#each metrics as metric}
		<div class="rounded-lg border bg-card p-4">
			<div class="text-xs font-medium uppercase tracking-wide text-muted-foreground">
				{metric.label}
			</div>
			<div class="mt-1 font-mono text-lg font-semibold tabular-nums">
				{metric.value != null && metric.value !== 0 ? metric.format(metric.value) : '-'}
			</div>
		</div>
	{/each}
</div>
