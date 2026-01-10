<script lang="ts">
	import type { Company, Price } from '$lib/gen/ntx/v1/common_pb';
	import Plus from '@lucide/svelte/icons/plus';

	interface Props {
		company: Company;
		price?: Price;
	}

	let { company, price }: Props = $props();

	let currentPrice = $derived(price?.ltp ?? price?.close);

	function fmt(value: number | undefined): string {
		if (value === undefined) return '—';
		return value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	let hasChange = $derived(price?.changePercent !== undefined && price?.changePercent !== null);
	let isPositive = $derived((price?.changePercent ?? 0) >= 0);
</script>

<header class="px-6 py-4">
	<!-- Add to Watchlist -->
	<button
		class="mb-4 inline-flex items-center gap-1.5 rounded border border-border px-3 py-1.5 text-sm transition-colors hover:bg-muted"
	>
		<Plus class="size-4" />
		Add To Watchlist
	</button>

	<!-- Company Info Row -->
	<div class="flex flex-wrap items-center gap-x-3 gap-y-2">
		<!-- Symbol -->
		<h1 class="font-serif text-3xl font-medium tracking-tight">{company.symbol}</h1>

		<!-- Company Name -->
		<span class="text-lg text-muted-foreground">{company.name}</span>

		<span class="text-muted-foreground">•</span>

		<!-- Exchange -->
		<span class="text-muted-foreground">NEPSE</span>

		<span class="text-muted-foreground">•</span>

		<!-- Price -->
		<span class="text-xl tabular-nums">Rs. {fmt(currentPrice)}</span>

		<!-- Change (only show if data available) -->
		{#if hasChange}
			<span class="tabular-nums {isPositive ? 'text-positive' : 'text-negative'}">
				{#if price.changePercent < 0}▼{:else if price.changePercent > 0}▲{/if}
				{price.changePercent > 0 ? '+' : ''}{fmt(price.changePercent)}%
				<span class="text-sm"
					>({price.change && price.change > 0 ? '+' : ''}{fmt(price.change)})</span
				>
			</span>
		{/if}
	</div>
</header>
