<script lang="ts">
	import { formatPriceCompact, formatChange } from '$lib/utils/format';
	import { TrendingUpIcon, TrendingDownIcon, ArrowRightIcon } from '@lucide/svelte';
	import type { Price } from '@ntx/api/ntx/v1/common_pb';

	let {
		title,
		stocks,
		type,
		href
	}: {
		title: string;
		stocks: Price[];
		type: 'gainers' | 'losers';
		href?: string;
	} = $props();

	const isGainer = $derived(type === 'gainers');
</script>

<div class="rounded-xl border bg-card">
	<div class="flex items-center justify-between border-b px-4 py-3">
		<div class="flex items-center gap-2">
			{#if isGainer}
				<TrendingUpIcon class="h-4 w-4 text-positive" />
			{:else}
				<TrendingDownIcon class="h-4 w-4 text-negative" />
			{/if}
			<h3 class="font-medium">{title}</h3>
		</div>
		{#if href}
			<a
				{href}
				class="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
			>
				View all
				<ArrowRightIcon class="h-3 w-3" />
			</a>
		{/if}
	</div>
	<div class="divide-y">
		{#each stocks.slice(0, 5) as stock}
			<a
				href="/company/{stock.symbol}"
				class="flex items-center justify-between px-4 py-3 transition-colors hover:bg-accent/50"
			>
				<div class="flex items-center gap-3">
					<div
						class="flex h-8 w-8 items-center justify-center rounded-md text-xs font-bold {isGainer
							? 'bg-positive-muted text-positive'
							: 'bg-negative-muted text-negative'}"
					>
						{stock.symbol.slice(0, 2)}
					</div>
					<span class="font-mono font-medium">{stock.symbol}</span>
				</div>
				<div class="text-right">
					<div class="font-mono text-sm tabular-nums">{formatPriceCompact(stock.ltp)}</div>
					<div
						class="text-xs font-medium tabular-nums {isGainer
							? 'text-positive'
							: 'text-negative'}"
					>
						{formatChange(stock.percentChange)}
					</div>
				</div>
			</a>
		{/each}
		{#if stocks.length === 0}
			<div class="px-4 py-8 text-center text-sm text-muted-foreground">No data available</div>
		{/if}
	</div>
</div>
