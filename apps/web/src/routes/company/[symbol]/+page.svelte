<script lang="ts">
	import {
		StoryOpening,
		ValuationStrips,
		TimeRangeSelector,
		StatsPanel,
		AboutSection,
		FinancialsTable,
		CorporateActionsTable,
		AIResearchButton,
		MutualFundHoldings,
		FiscalYears,
		OwnershipHistory,
		type ViewMode
	} from '#lib/components/company/index.js';
	import {
		PriceChart,
		VolumeChart,
		EarningsChart,
		DividendChart,
		RatingsRadar,
		SectorPeers
	} from '#lib/components/charts/index.js';

	let { data } = $props();

	let company = $derived(data.company);
	let story = $derived(data.story);
	let fundamentals = $derived(data.fundamentals);
	let fundamentalsHistory = $derived(data.fundamentalsHistory ?? []);
	let priceData = $derived(data.price);
	let priceHistory = $derived(data.priceHistory);
	let sectorStats = $derived(data.sectorStats);
	let ownership = $derived(data.ownership);
	let corporateActions = $derived(data.corporateActions ?? []);
	let fundHoldings = $derived(data.fundHoldings ?? []);

	let chartDays = $state<number>(365);
	let viewMode = $state<ViewMode>('quarterly');

	function handleDaysChange(days: number) {
		chartDays = days;
	}

	function handleViewModeChange(mode: ViewMode) {
		viewMode = mode;
	}

	let filteredFundamentals = $derived(
		fundamentalsHistory.filter((f) => (viewMode === 'quarterly' ? !!f.quarter : !f.quarter))
	);
</script>

{#if company}
	<div class="min-h-screen bg-background text-foreground">
		<div class="mx-auto max-w-7xl px-4 py-8">
			<!-- Opening: what the filings say, then where the ratios sit -->
			<StoryOpening {company} {story} />
			<ValuationStrips strips={story?.strips ?? []} />

			<!-- One row per fiscal year, newest first, with compounded growth beside it -->
			<div class="mt-12 border-t border-border pt-8">
				<FiscalYears years={story?.fiscalYears ?? []} growth={story?.growth ?? []} />
			</div>

			<!-- Main Content Grid -->
			<div class="mt-8 grid gap-8 lg:grid-cols-3">
				<!-- Chart: span 2 columns -->
				<div class="min-w-0 lg:col-span-2">
					<div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
						<h2 class="font-serif text-lg font-medium">Price History</h2>
						<TimeRangeSelector selected={chartDays} onSelect={handleDaysChange} />
					</div>
					<div class="h-[350px] overflow-hidden">
						<PriceChart prices={priceHistory} events={story?.events ?? []} days={chartDays} />
					</div>
					<div class="mt-2">
						<h3 class="mb-1 text-xs font-medium text-muted-foreground">Volume</h3>
						<VolumeChart prices={priceHistory} days={chartDays} />
					</div>
				</div>

				<!-- Stats + AI: single column, stacked -->
				<div class="flex min-w-0 flex-col gap-6">
					<StatsPanel
						price={priceData}
						{fundamentals}
						{priceHistory}
						{ownership}
						{corporateActions}
					/>
					<AIResearchButton {company} price={priceData} {fundamentals} {sectorStats} />
				</div>

				<!-- About: below chart, span 2 columns -->
				<div class="min-w-0 lg:col-span-2">
					<AboutSection {company} />
				</div>
			</div>

			<!-- Financial History & Rating -->
			<div class="mt-12 border-t border-border pt-8">
				<div class="grid gap-8 lg:grid-cols-12">
					<div class="min-w-0 lg:col-span-8">
						<FinancialsTable
							fundamentals={fundamentalsHistory}
							{viewMode}
							onViewModeChange={handleViewModeChange}
						/>
					</div>
					<div class="min-w-0 lg:col-span-4">
						<RatingsRadar fundamentals={filteredFundamentals} price={priceData} />
					</div>
				</div>
			</div>

			<!-- Earnings & Ownership -->
			<div class="mt-12 border-t border-border pt-8">
				<div class="grid gap-8 lg:grid-cols-12">
					<div id="quarterly" class="min-w-0 lg:col-span-8">
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
						<EarningsChart quarters={story?.quarters ?? []} />
					</div>
					<div class="min-w-0 lg:col-span-4">
						<OwnershipHistory observations={story?.ownership ?? []} />
					</div>
				</div>
			</div>

			<!-- Dividends -->
			<div class="mt-12 border-t border-border pt-8">
				<div class="grid gap-8 lg:grid-cols-12">
					<div class="order-2 min-w-0 lg:order-1 lg:col-span-4">
						<CorporateActionsTable actions={corporateActions} />
					</div>
					<div class="order-1 min-w-0 lg:order-2 lg:col-span-8">
						<h3 class="mb-4 font-serif text-base font-medium">Dividends</h3>
						<DividendChart actions={corporateActions} />
					</div>
				</div>
			</div>

			<!-- Mutual Fund Holdings -->
			{#if fundHoldings.length > 0}
				<div id="funds" class="mt-12 border-t border-border pt-8">
					<MutualFundHoldings holdings={fundHoldings} />
				</div>
			{/if}

			<!-- Sector Peers -->
			<div class="mt-12 border-t border-border pt-8">
				<SectorPeers {company} allCompanies={data.companies ?? []} allPrices={data.prices ?? []} />
			</div>
		</div>
	</div>
{/if}
