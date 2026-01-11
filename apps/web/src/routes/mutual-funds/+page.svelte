<script lang="ts">
	import { FundCard } from '$lib/components/mutual-funds';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import TrendingUp from '@lucide/svelte/icons/trending-up';
	import TrendingDown from '@lucide/svelte/icons/trending-down';
	import Wallet from '@lucide/svelte/icons/wallet';

	let { data } = $props();

	// Calculate totals
	let totalAUM = $derived(data.funds.reduce((sum, f) => sum + f.net_assets, 0));

	// Funds above/below par
	let abovePar = $derived(data.funds.filter((f) => f.nav_per_unit > 10));
	let belowPar = $derived(data.funds.filter((f) => f.nav_per_unit < 10));

	// Best and worst performing
	let bestFund = $derived([...data.funds].sort((a, b) => b.nav_per_unit - a.nav_per_unit)[0]);
	let worstFund = $derived([...data.funds].sort((a, b) => a.nav_per_unit - b.nav_per_unit)[0]);

	function fmtLarge(value: number): string {
		if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`;
		if (value >= 10_000_000) return `${(value / 10_000_000).toFixed(2)} Cr`;
		return value.toLocaleString('en-NP');
	}

	// Report date from first fund
	let reportDate = $derived(data.funds[0]?.report_date_nepali ?? 'N/A');
</script>

<svelte:head>
	<title>Mutual Funds | NTX</title>
	<meta name="description" content="NEPSE Open-End Mutual Funds NAV and Portfolio Analysis" />
</svelte:head>

<div class="min-h-screen">
	<!-- Header -->
	<header class="border-b border-border">
		<div class="mx-auto max-w-7xl px-4">
			<!-- Top bar -->
			<div class="flex items-center justify-between py-3 text-xs text-muted-foreground">
				<a href="/" class="flex items-center gap-2 transition-colors hover:text-foreground">
					<ArrowLeft class="size-4" />
					<span>Back to Market</span>
				</a>
				<div class="flex items-center gap-3">
					<span>Data as of {reportDate}</span>
					<ThemeToggle />
				</div>
			</div>

			<!-- Title -->
			<div class="border-t border-border py-8">
				<h1 class="font-serif text-4xl tracking-tight md:text-5xl">Mutual Funds</h1>
				<p class="mt-2 text-muted-foreground">Open-End Mutual Fund NAV and Portfolio Analysis</p>
			</div>
		</div>
	</header>

	<!-- Main Content -->
	<main class="mx-auto max-w-7xl px-4 py-8">
		<!-- Summary Cards -->
		<div class="mb-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<!-- Total AUM -->
			<div class="rounded-xl border border-border bg-card/50 p-4">
				<div class="flex items-center gap-2 text-muted-foreground">
					<Wallet class="size-4" />
					<span class="text-sm">Total AUM</span>
				</div>
				<p class="mt-2 font-serif text-2xl tabular-nums">{fmtLarge(totalAUM)}</p>
				<p class="text-xs text-muted-foreground">{data.funds.length} funds</p>
			</div>

			<!-- Above Par -->
			<div class="rounded-xl border border-border bg-card/50 p-4">
				<div class="flex items-center gap-2 text-positive">
					<TrendingUp class="size-4" />
					<span class="text-sm">Above Par</span>
				</div>
				<p class="mt-2 font-serif text-2xl tabular-nums">{abovePar.length}</p>
				<p class="text-xs text-muted-foreground">NAV &gt; 10</p>
			</div>

			<!-- Below Par -->
			<div class="rounded-xl border border-border bg-card/50 p-4">
				<div class="flex items-center gap-2 text-negative">
					<TrendingDown class="size-4" />
					<span class="text-sm">Below Par</span>
				</div>
				<p class="mt-2 font-serif text-2xl tabular-nums">{belowPar.length}</p>
				<p class="text-xs text-muted-foreground">NAV &lt; 10</p>
			</div>

			<!-- Best Performer -->
			<div class="rounded-xl border border-border bg-card/50 p-4">
				<div class="flex items-center gap-2 text-muted-foreground">
					<span class="text-sm">Top Performer</span>
				</div>
				<p class="mt-2 font-serif text-2xl">{bestFund?.symbol}</p>
				<p class="text-xs text-positive tabular-nums">NAV {bestFund?.nav_per_unit.toFixed(2)}</p>
			</div>
		</div>

		<!-- Fund Grid -->
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each data.funds as fund (fund.symbol)}
				<FundCard {fund} />
			{/each}
		</div>
	</main>
</div>
