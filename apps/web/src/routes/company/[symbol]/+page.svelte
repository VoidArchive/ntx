<script lang="ts">
	import {
		StoryOpening,
		ValuationStrips,
		TimeRangeSelector,
		FinancialsTable,
		AIResearchButton,
		MutualFundHoldings,
		FiscalYears,
		OwnershipHistory,
		Peers
	} from '#lib/components/company/index.js';
	import { PriceChart, VolumeChart, EarningsChart } from '#lib/components/charts/index.js';

	let { data } = $props();

	let company = $derived(data.company);
	let story = $derived(data.story);
	let fundamentalsHistory = $derived(data.fundamentalsHistory ?? []);
	let priceHistory = $derived(data.priceHistory ?? []);
	let fundHoldings = $derived(data.fundHoldings ?? []);
	let lastFiling = $derived(
		story?.lastFiling ? `Last filing: ${story.lastFiling}.` : 'No filings recorded.'
	);
	let filedDaysAgo = $derived(
		story?.filedAt ? Math.floor((Date.now() - Date.parse(story.filedAt)) / 86_400_000) : undefined
	);

	let chartDays = $state<number>(365);
</script>

<!-- A section without data keeps its anchor and collapses to one line. -->
{#snippet empty(id: string, title: string, text: string)}
	<section {id} class="mt-8 border-t border-border pt-4 text-sm">
		<span class="font-serif font-medium">{title}</span>
		<span class="text-muted-foreground">· {text}</span>
	</section>
{/snippet}

{#if company}
	<div class="min-h-screen bg-background text-foreground">
		<div class="mx-auto max-w-7xl px-4 py-8">
			<!-- 1. What the filings say, checks beside it, price line below -->
			<StoryOpening {company} {story} />

			<!-- 2. Where the ratios sit -->
			{#if story?.strips.length}
				<ValuationStrips strips={story.strips} />
			{:else}
				{@render empty('strips', 'Valuation', `No ratios recorded. ${lastFiling}`)}
			{/if}

			<!-- 3. Price and volume, with results and notices marked -->
			{#if priceHistory.length > 0}
				<section id="price" class="mt-12 border-t border-border pt-8">
					<div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
						<h2 class="font-serif text-lg font-medium">Price and volume</h2>
						<TimeRangeSelector selected={chartDays} onSelect={(days) => (chartDays = days)} />
					</div>
					<div class="h-[350px] overflow-hidden">
						<PriceChart prices={priceHistory} events={story?.events ?? []} days={chartDays} />
					</div>
					<div class="mt-2">
						<VolumeChart prices={priceHistory} days={chartDays} />
					</div>
				</section>
			{:else}
				{@render empty('price', 'Price and volume', 'No closes recorded.')}
			{/if}

			<!-- 4. One row per fiscal year with compounded growth beside it -->
			{#if story?.fiscalYears.length}
				<div class="mt-12 border-t border-border pt-8">
					<FiscalYears years={story.fiscalYears} growth={story.growth} />
				</div>
			{:else}
				{@render empty(
					'fiscal-years',
					'Fiscal years',
					`No full-year filings or notices. ${lastFiling}`
				)}
			{/if}

			<!-- 5. Quarter-only results, then the rows as filed -->
			{#if story?.quarters.length || fundamentalsHistory.some((f) => !!f.quarter)}
				<section id="quarterly" class="mt-12 border-t border-border pt-8">
					<div class="mb-4 flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
						<h3 class="font-serif text-base font-medium">Quarterly results</h3>
						{#if story?.hydroEps}
							{@const h = story.hydroEps}
							<p class="text-xs text-muted-foreground tabular-nums">
								EPS Rs {h.quarter.toFixed(2)} quarter · Rs {h.annualized.toFixed(2)} annualized · Rs
								{h.fullYear.toFixed(2)}
								{h.fullYearPeriod}
							</p>
						{/if}
					</div>
					{#if story?.quarters.length}
						<EarningsChart quarters={story.quarters} />
					{/if}
					<FinancialsTable fundamentals={fundamentalsHistory} class="mt-6" />
				</section>
			{:else}
				{@render empty('quarterly', 'Quarterly results', `No quarterly filings. ${lastFiling}`)}
			{/if}

			<!-- 6. Ownership over time and fund holdings -->
			{#if story?.ownership.length}
				<div class="mt-12 border-t border-border pt-8">
					<OwnershipHistory observations={story.ownership} />
				</div>
			{:else}
				{@render empty('ownership', 'Ownership', 'No shareholding observations recorded.')}
			{/if}
			{#if fundHoldings.length > 0}
				<div id="funds" class="mt-12 border-t border-border pt-8">
					<MutualFundHoldings holdings={fundHoldings} />
				</div>
			{:else}
				{@render empty(
					'funds',
					'Fund holdings',
					'No open-ended fund lists it in the NAV reports on file.'
				)}
			{/if}

			<!-- 7. Same sector by PE, on the PE strip's sector range -->
			{#if story?.peers.length}
				<div class="mt-12 border-t border-border pt-8">
					<Peers
						peers={story.peers}
						range={story.strips.find((s) => s.ratio === 'pe')?.sector}
						symbol={company.symbol}
					/>
				</div>
			{:else}
				{@render empty('peers', 'Peers', 'No sector company has a quarterly filing and a close.')}
			{/if}

			<!-- 8. How fresh the data is, and the research prompt built from it -->
			<footer
				class="mt-12 flex flex-col gap-3 border-t border-border pt-6 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between"
			>
				<p class="tabular-nums">
					Last sync {story?.syncedAt ? `${story.syncedAt} UTC` : 'not recorded'} · Newest filing
					{#if story?.filedAt}
						{story.lastFiling}, {story.filedAtSeen ? 'first seen by NTX on' : 'published'}
						{story.filedAt} ({filedDaysAgo} days ago)
					{:else}
						none
					{/if}
					·
					<a href="#quarterly" class="underline underline-offset-2 hover:text-foreground"
						>raw rows</a
					>
				</p>
				<AIResearchButton {company} {story} />
			</footer>
		</div>
	</div>
{/if}
