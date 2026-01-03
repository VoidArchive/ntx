<script lang="ts">
	import { formatPriceCompact, formatChange } from '$lib/utils/format';
	import { TrendingUpIcon, TrendingDownIcon, MinusIcon } from '@lucide/svelte';

	let {
		symbol,
		name,
		price,
		change,
		percentChange,
		href
	}: {
		symbol: string;
		name?: string;
		price: number;
		change: number;
		percentChange: number;
		href?: string;
	} = $props();

	const isPositive = change > 0;
	const isNegative = change < 0;
</script>

<a
	{href}
	class="group flex items-center justify-between rounded-lg border bg-card p-4 transition-colors hover:bg-accent/50"
>
	<div class="flex items-center gap-3">
		<div
			class="flex h-10 w-10 items-center justify-center rounded-md text-sm font-bold {isPositive
				? 'bg-positive-muted text-positive'
				: isNegative
					? 'bg-negative-muted text-negative'
					: 'bg-muted text-muted-foreground'}"
		>
			{symbol.slice(0, 2)}
		</div>
		<div>
			<div class="font-mono font-semibold group-hover:text-primary">{symbol}</div>
			{#if name}
				<div class="text-xs text-muted-foreground line-clamp-1">{name}</div>
			{/if}
		</div>
	</div>
	<div class="text-right">
		<div class="font-mono text-sm font-medium tabular-nums">
			{formatPriceCompact(price)}
		</div>
		<div
			class="flex items-center justify-end gap-1 text-xs {isPositive
				? 'text-positive'
				: isNegative
					? 'text-negative'
					: 'text-neutral'}"
		>
			{#if isPositive}
				<TrendingUpIcon class="h-3 w-3" />
			{:else if isNegative}
				<TrendingDownIcon class="h-3 w-3" />
			{:else}
				<MinusIcon class="h-3 w-3" />
			{/if}
			<span class="font-medium tabular-nums">{formatChange(percentChange)}</span>
		</div>
	</div>
</a>
