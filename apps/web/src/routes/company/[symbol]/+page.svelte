<script lang="ts">
	import {
		CompanyHeader,
		TimeRangeSelector,
		StatsPanel,
		AboutSection,
		FinancialsTable,
		CorporateActionsTable,
		AIResearchButton,
		MutualFundHoldings,
		type ViewMode
	} from '$lib/components/company';
	import {
		PriceChart,
		VolumeChart,
		EarningsChart,
		DividendChart,
		OwnershipPieChart,
		RatingsRadar
	} from '$lib/components/charts';

	let { data } = $props();

	let company = $derived(data.company);
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
			<!-- Header -->
			<CompanyHeader {company} price={priceData} />

			<!-- Main Content Grid -->
			<div class="mt-8 grid gap-8 lg:grid-cols-3">
				<!-- Chart: span 2 columns -->
				<div class="min-w-0 lg:col-span-2">
					<div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
						<h2 class="font-serif text-lg font-medium">Price History</h2>
						<TimeRangeSelector selected={chartDays} onSelect={handleDaysChange} />
					</div>
					<div class="h-[350px] overflow-hidden">
						<PriceChart prices={priceHistory} days={chartDays} />
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
					<div class="min-w-0 lg:col-span-8">
						<h3 class="mb-4 font-serif text-base font-medium">Earnings</h3>
						<EarningsChart fundamentals={filteredFundamentals} />
					</div>
					<div class="min-w-0 lg:col-span-4">
						<OwnershipPieChart {ownership} />
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
				<div class="mt-12 border-t border-border pt-8">
					<MutualFundHoldings holdings={fundHoldings} />
				</div>
			{/if}
		</div>
	</div>
{/if}
