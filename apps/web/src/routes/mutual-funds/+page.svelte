<script lang="ts">
	import { formatLarge } from '#lib/utils.js';
	import { FundCard, fundProvenance } from '#lib/components/mutual-funds/index.js';
	import LayoutGrid from '@lucide/svelte/icons/layout-grid';
	import List from '@lucide/svelte/icons/list';
	import type { Fund } from '#lib/types/fund.js';

	let { data } = $props();

	let viewMode = $state<'grid' | 'table'>('table');

	// Calculate totals
	let totalAUM = $derived(data.funds.reduce((sum: number, f: Fund) => sum + f.net_assets, 0));

	// Funds above/below par
	let abovePar = $derived(data.funds.filter((f) => f.nav_per_unit > 10));
	let belowPar = $derived(data.funds.filter((f) => f.nav_per_unit < 10));

	// Best performer
	let bestFund = $derived([...data.funds].sort((a, b) => b.nav_per_unit - a.nav_per_unit)[0]);

	// Every fund in the snapshot shares one report date, so the first fund speaks for all
	let provenance = $derived(data.funds[0] ? fundProvenance(data.funds[0]) : '');
</script>

<svelte:head>
	<title>Mutual Funds | NTX</title>
	<meta name="description" content="NEPSE Open-End Mutual Funds NAV and Portfolio Analysis" />
</svelte:head>

<div class="min-h-screen">
	<!-- Main Content -->
	<main class="mx-auto max-w-7xl px-4 py-8">
		<!-- Page Header -->
		<div class="mb-8 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
			<div>
				<h1 class="font-serif text-3xl font-medium tracking-tight">Mutual funds</h1>
				<!-- The sentence first, as on every page. -->
				<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">
					{data.funds.length} open-end funds hold Rs {formatLarge(totalAUM)} in net assets; {abovePar.length}
					are above the Rs 10 par and {belowPar.length} below. The highest NAV is {bestFund?.symbol}
					at Rs {bestFund?.nav_per_unit.toFixed(2)}.
				</p>
				<p class="mt-2 text-sm text-muted-foreground">{provenance}</p>
			</div>

			<!-- View Toggle -->
			<div class="flex items-center gap-1 rounded-lg border border-border bg-muted/30 p-1">
				<button
					onclick={() => (viewMode = 'grid')}
					class="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm transition-colors {viewMode ===
					'grid'
						? 'bg-background text-foreground shadow-sm'
						: 'text-muted-foreground hover:text-foreground'}"
				>
					<LayoutGrid class="size-4" />
					<span class="hidden sm:inline">Grid</span>
				</button>
				<button
					onclick={() => (viewMode = 'table')}
					class="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm transition-colors {viewMode ===
					'table'
						? 'bg-background text-foreground shadow-sm'
						: 'text-muted-foreground hover:text-foreground'}"
				>
					<List class="size-4" />
					<span class="hidden sm:inline">Table</span>
				</button>
			</div>
		</div>
		<!-- Fund Grid -->
		{#if viewMode === 'grid'}
			<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
				{#each data.funds as fund (fund.symbol)}
					<FundCard {fund} />
				{/each}
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-sm whitespace-nowrap tabular-nums [&_td+td]:pl-3 [&_th+th]:pl-3">
					<thead class="text-xs text-muted-foreground">
						<tr class="border-b border-border">
							<th class="py-2 text-left font-normal">Fund</th>
							<th class="py-2 text-left font-normal">Name</th>
							<th class="py-2 text-right font-normal">NAV</th>
							<th class="py-2 text-right font-normal">From par</th>
							<th class="py-2 text-right font-normal">Net assets</th>
							<th class="py-2 text-left font-normal">Holdings</th>
							<th class="py-2 text-left font-normal">Report</th>
						</tr>
					</thead>
					<tbody>
						{#each data.funds as fund (fund.symbol)}
							{@const fromPar = ((fund.nav_per_unit - 10) / 10) * 100}
							<tr class="border-b border-border/50 hover:bg-muted/50">
								<td class="py-1.5">
									<a
										href="/mutual-funds/{fund.symbol}"
										class="font-serif font-medium hover:underline">{fund.symbol}</a
									>
								</td>
								<td class="max-w-56 truncate py-1.5 text-muted-foreground" title={fund.fund_name}
									>{fund.fund_name}</td
								>
								<td class="py-1.5 text-right">{fund.nav_per_unit.toFixed(2)}</td>
								<!-- Sign and glyph carry direction, so colour is never the only signal. -->
								<td
									class="py-1.5 text-right {fromPar > 0
										? 'text-positive'
										: fromPar < 0
											? 'text-negative'
											: 'text-muted-foreground'}"
								>
									{fromPar > 0 ? '▲' : fromPar < 0 ? '▼' : '■'}
									{Math.abs(fromPar).toFixed(1)}%
								</td>
								<td class="py-1.5 text-right">{formatLarge(fund.net_assets)}</td>
								<td class="py-1.5 text-muted-foreground">
									{#if fund.holdings_disclosure === 'full'}
										itemised
									{:else if fund.holdings_disclosure === 'top10'}
										top 10 only
									{:else}
										not published
									{/if}
								</td>
								<td class="py-1.5 text-xs text-muted-foreground">{fund.report_date_nepali}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</main>
</div>
