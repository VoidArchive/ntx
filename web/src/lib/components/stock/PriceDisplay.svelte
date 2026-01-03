<script lang="ts">
	import { formatPriceCompact, formatChange } from '$lib/utils/format';
	import { TrendingUpIcon, TrendingDownIcon, MinusIcon } from '@lucide/svelte';

	let {
		price,
		change,
		percentChange,
		size = 'default'
	}: {
		price: number;
		change: number;
		percentChange: number;
		size?: 'sm' | 'default' | 'lg';
	} = $props();

	const isPositive = $derived(change > 0);
	const isNegative = $derived(change < 0);

	const sizeClasses = {
		sm: 'text-base',
		default: 'text-2xl',
		lg: 'text-4xl'
	};

	const changeSizeClasses = {
		sm: 'text-xs',
		default: 'text-sm',
		lg: 'text-base'
	};

	const iconSizes = {
		sm: 'h-3 w-3',
		default: 'h-4 w-4',
		lg: 'h-5 w-5'
	};
</script>

<div class="flex flex-col gap-1">
	<span class="font-mono font-semibold tabular-nums {sizeClasses[size]}">
		Rs. {formatPriceCompact(price)}
	</span>
	<div
		class="flex items-center gap-1 font-medium {changeSizeClasses[size]} {isPositive
			? 'text-positive'
			: isNegative
				? 'text-negative'
				: 'text-neutral'}"
	>
		{#if isPositive}
			<TrendingUpIcon class={iconSizes[size]} />
		{:else if isNegative}
			<TrendingDownIcon class={iconSizes[size]} />
		{:else}
			<MinusIcon class={iconSizes[size]} />
		{/if}
		<span class="tabular-nums">
			{formatChange(percentChange)}
		</span>
		<span class="tabular-nums text-muted-foreground">
			({change >= 0 ? '+' : ''}{change.toFixed(2)})
		</span>
	</div>
</div>
