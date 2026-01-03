<script lang="ts">
	import { formatNumber, formatChange } from '$lib/utils/format';
	import { TrendingUpIcon, TrendingDownIcon, MinusIcon } from '@lucide/svelte';

	let {
		name,
		value,
		change,
		percentChange
	}: {
		name: string;
		value: number;
		change: number;
		percentChange: number;
	} = $props();

	const isPositive = change > 0;
	const isNegative = change < 0;
</script>

<div class="rounded-xl border bg-card p-4">
	<div class="text-xs font-medium text-muted-foreground">{name}</div>
	<div class="mt-1 font-mono text-2xl font-bold tabular-nums">
		{formatNumber(value, 2)}
	</div>
	<div
		class="mt-1 flex items-center gap-1 text-sm font-medium {isPositive
			? 'text-positive'
			: isNegative
				? 'text-negative'
				: 'text-neutral'}"
	>
		{#if isPositive}
			<TrendingUpIcon class="h-4 w-4" />
		{:else if isNegative}
			<TrendingDownIcon class="h-4 w-4" />
		{:else}
			<MinusIcon class="h-4 w-4" />
		{/if}
		<span class="tabular-nums">{formatChange(percentChange)}</span>
		<span class="text-xs text-muted-foreground tabular-nums">
			({change >= 0 ? '+' : ''}{change.toFixed(2)})
		</span>
	</div>
</div>
