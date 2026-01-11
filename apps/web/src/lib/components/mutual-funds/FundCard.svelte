<script lang="ts">
	import type { Fund } from '$lib/types/fund';
	import ArrowRight from '@lucide/svelte/icons/arrow-right';
	import Wallet from '@lucide/svelte/icons/wallet';

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

	// Count total holdings across all sectors
	let holdingsCount = $derived.by(() => {
		let count = 0;
		for (const holdings of Object.values(fund.holdings)) {
			if (Array.isArray(holdings)) {
				count += holdings.length;
			}
		}
		return count;
	});

	// Calculate if NAV is above or below par (10)
	let navStatus = $derived.by(() => {
		if (fund.nav_per_unit > 10) return 'positive';
		if (fund.nav_per_unit < 10) return 'negative';
		return 'neutral';
	});
</script>

<a
	href="/mutual-funds/{fund.symbol}"
	class="group relative flex flex-col justify-between overflow-hidden rounded-xl border border-border bg-card/50 p-6 shadow-sm backdrop-blur-sm transition-all hover:-translate-y-1 hover:border-foreground/20 hover:shadow-md"
>
	<!-- Top: Header -->
	<div class="flex items-start justify-between gap-4">
		<div class="min-w-0 flex-1">
			<h3 class="font-serif text-2xl font-medium tracking-tight text-foreground group-hover:underline">
				{fund.symbol}
			</h3>
			<p class="mt-1 line-clamp-2 text-sm text-muted-foreground" title={fund.fund_name}>
				{fund.fund_name}
			</p>
		</div>

		<!-- Fund Size Badge -->
		<div class="flex shrink-0 items-center gap-1.5 rounded-full border border-border bg-background/50 px-2.5 py-1">
			<Wallet class="size-3 text-muted-foreground" />
			<span class="text-xs font-medium tabular-nums">{fmtLarge(fund.net_assets)}</span>
		</div>
	</div>

	<!-- Middle: NAV -->
	<div class="mt-6 flex items-baseline gap-3">
		<span class="text-3xl font-medium text-foreground tabular-nums">
			{fund.nav_per_unit.toFixed(2)}
		</span>
		<span class="text-sm text-muted-foreground">NAV/Unit</span>
		{#if navStatus !== 'neutral'}
			<span
				class="ml-auto text-sm font-medium tabular-nums {navStatus === 'positive'
					? 'text-positive'
					: 'text-negative'}"
			>
				{navStatus === 'positive' ? '+' : ''}{((fund.nav_per_unit - 10) / 10 * 100).toFixed(1)}%
			</span>
		{/if}
	</div>

	<!-- Stats Row -->
	<div class="mt-4 flex items-center gap-4 text-xs text-muted-foreground">
		<span>{holdingsCount} holdings</span>
		<span class="text-border">|</span>
		<span>{fund.report_date_nepali}</span>
	</div>

	<!-- Bottom: CTA -->
	<div
		class="mt-6 flex items-center justify-between border-t border-border/50 pt-4 opacity-60 transition-opacity group-hover:opacity-100"
	>
		<span class="text-xs font-medium text-muted-foreground">View Portfolio</span>
		<ArrowRight class="size-4 text-muted-foreground transition-transform group-hover:translate-x-1" />
	</div>

	<!-- Hover Gradient -->
	<div
		class="absolute inset-0 -z-10 bg-gradient-to-br from-primary/5 via-transparent to-transparent opacity-0 transition-opacity group-hover:opacity-100"
	></div>
</a>
