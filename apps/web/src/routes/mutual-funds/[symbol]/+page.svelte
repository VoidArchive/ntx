<script lang="ts">
	import {
		FundHeader,
		FundStats,
		PortfolioDonut,
		HoldingsTable,
		fundOpeningSentence,
		fundProvenance
	} from '#lib/components/mutual-funds/index.js';

	let { data } = $props();

	let fund = $derived(data.fund);
</script>

<svelte:head>
	<title>{fund.symbol} - {fund.fund_name} | NTX</title>
	<meta name="description" content="{fund.fund_name} NAV and Portfolio Holdings" />
</svelte:head>

<div class="min-h-screen bg-background text-foreground">
	<div class="mx-auto max-w-7xl px-4 py-8">
		<!-- Header -->
		<FundHeader {fund} />

		<p class="mt-6 max-w-3xl text-lg leading-relaxed">{fundOpeningSentence(fund)}</p>
		<p class="mt-2 text-sm text-muted-foreground">{fundProvenance(fund)}</p>

		<!-- Main Content Grid -->
		<div class="mt-8 grid gap-8 lg:grid-cols-3">
			<!-- Portfolio Chart: span 2 columns -->
			<div class="min-w-0 lg:col-span-2">
				<PortfolioDonut holdings={fund.holdings} />
			</div>

			<!-- Stats Panel -->
			<div class="min-w-0">
				<h3 class="mb-4 font-serif text-lg font-medium">Fund Statistics</h3>
				<FundStats {fund} />
			</div>
		</div>

		<!-- Holdings Table -->
		<div class="mt-8">
			<HoldingsTable holdings={fund.holdings} netAssets={fund.net_assets} />
		</div>

		<!-- Footer note -->
		<div class="mt-8 text-center text-xs text-muted-foreground">
			<p>Managed by {fund.fund_manager}</p>
			{#if fund.holdings_disclosure === 'top10'}
				<p class="mt-1">
					Only the 10 largest holdings are published; the rest of the share portfolio is shown as
					"Other listed shares".
				</p>
			{:else if fund.holdings_disclosure === 'none'}
				<p class="mt-1">This fund does not publish itemised holdings.</p>
			{/if}
			<p class="mt-1">
				Allocation figures are gross of liabilities, so they can exceed net assets.
			</p>
		</div>
	</div>
</div>
