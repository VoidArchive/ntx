<script lang="ts">
	import {
		FundHeader,
		FundStats,
		PortfolioDonut,
		HoldingsTable
	} from '$lib/components/mutual-funds';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	let { data } = $props();

	let fund = $derived(data.fund);
</script>

<svelte:head>
	<title>{fund.symbol} - {fund.fund_name} | NTX</title>
	<meta name="description" content="{fund.fund_name} NAV and Portfolio Holdings" />
</svelte:head>

<div class="min-h-screen bg-background text-foreground">
	<!-- Sticky Header -->
	<div class="sticky top-0 z-50 border-b border-border bg-background/80 backdrop-blur-sm">
		<div class="mx-auto flex max-w-7xl items-center justify-between px-4 py-3">
			<span class="font-serif text-lg">{fund.symbol}</span>
			<div class="flex items-center gap-4">
				<span class="text-sm tabular-nums">NAV {fund.nav_per_unit.toFixed(2)}</span>
				<ThemeToggle />
			</div>
		</div>
	</div>

	<div class="mx-auto max-w-7xl px-4 py-8">
		<!-- Header -->
		<FundHeader {fund} />

		<!-- Main Content Grid -->
		<div class="mt-8 grid gap-8 lg:grid-cols-3">
			<!-- Portfolio Chart: span 2 columns -->
			<div class="min-w-0 lg:col-span-2">
				<div class="rounded-xl border border-border bg-card/50 p-6">
					<PortfolioDonut holdings={fund.holdings} netAssets={fund.net_assets} />
				</div>
			</div>

			<!-- Stats Panel -->
			<div class="min-w-0">
				<div class="rounded-xl border border-border bg-card/50 p-6">
					<h3 class="mb-4 font-serif text-lg font-medium">Fund Statistics</h3>
					<FundStats {fund} />
				</div>
			</div>
		</div>

		<!-- Holdings Table -->
		<div class="mt-8">
			<div class="rounded-xl border border-border bg-card/50 p-6">
				<HoldingsTable holdings={fund.holdings} netAssets={fund.net_assets} />
			</div>
		</div>

		<!-- Footer note -->
		<div class="mt-8 text-center text-xs text-muted-foreground">
			<p>Data as of {fund.report_date_english} ({fund.report_date_nepali})</p>
			<p class="mt-1">Managed by {fund.fund_manager}</p>
		</div>
	</div>
</div>
