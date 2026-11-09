<script lang="ts">
	import { FundHeader, HoldingsTable } from '#lib/components/mutual-funds/index.js';

	let { data } = $props();

	let fund = $derived(data.fund);
</script>

<svelte:head>
	<title>{fund.symbol} - {fund.fund_name} | NTX</title>
	<meta name="description" content="{fund.fund_name} NAV and Portfolio Holdings" />
</svelte:head>

<div class="mx-auto max-w-7xl px-4 py-6">
	<FundHeader {fund} />

	<section class="mt-6 card-report">
		<HoldingsTable holdings={fund.holdings} netAssets={fund.net_assets} />
		<p class="mt-6 text-xs text-muted-foreground">
			{#if fund.holdings_disclosure === 'top10'}
				Only the 10 largest holdings are published; the rest of the share portfolio is shown as
				"Other listed shares".
			{:else if fund.holdings_disclosure === 'none'}
				This fund does not publish itemised holdings.
			{/if}
			Allocation figures are gross of liabilities, so they can exceed net assets.
		</p>
	</section>
</div>
