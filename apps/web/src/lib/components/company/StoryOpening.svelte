<script lang="ts">
	import type { Company, Price } from '#lib/gen/ntx/v1/common_pb.js';
	import type { GetCompanyStoryResponse } from '#lib/gen/ntx/v1/story_pb.js';
	import { sectorNames } from '#lib/utils.js';
	import { PriceChart } from '#lib/components/charts/index.js';
	import AIResearchButton from './AIResearchButton.svelte';
	import TimeRangeSelector from './TimeRangeSelector.svelte';

	interface Props {
		company: Company;
		story?: GetCompanyStoryResponse;
		prices?: Price[]; // the last year of closes, any order
		shares?: number; // listed shares, for the market cap
	}

	let { company, story, prices = [], shares = 0 }: Props = $props();

	let quote = $derived(story?.quote);
	let rating = $derived(story?.rating);
	let sector = $derived(sectorNames[company.sector] ?? 'Others');
	// Close times listed shares, the same arithmetic the screener uses.
	let marketCap = $derived(quote && shares > 0 ? shares * quote.close : undefined);
	let initials = $derived(
		company.name
			.split(/\s+/)
			.filter((w) => /^[A-Za-z]/.test(w))
			.slice(0, 2)
			.map((w) => w[0])
			.join('')
			.toUpperCase()
	);

	let series = $derived(
		[...prices]
			.sort((a, b) => a.businessDate.localeCompare(b.businessDate))
			.map((p) => p.close ?? p.ltp ?? 0)
			.filter((v) => v > 0)
	);
	let yearChange = $derived(
		series.length > 1 ? (series[series.length - 1] / series[0] - 1) * 100 : undefined
	);
	let chartDays = $state(365);

	function fmt(value: number): string {
		return value.toLocaleString('en-NP', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
	}
	function signed(value: number): string {
		return `${value > 0 ? '+' : ''}${fmt(value)}`;
	}
	function tone(value: number): string {
		return value > 0 ? 'text-positive' : value < 0 ? 'text-negative' : 'text-primary';
	}
</script>

<header class="rounded-lg bg-card p-6 shadow-card sm:p-8">
	<div class="grid gap-8 lg:grid-cols-[1fr_minmax(0,480px)]">
		<div class="min-w-0">
			<p class="text-sm text-muted-foreground">
				<a href="/company" class="hover:text-foreground">Stocks</a>
				<span aria-hidden="true" class="mx-1">/</span>
				<a href="/company?sector={company.sector}" class="hover:text-foreground">{sector}</a>
			</p>
			<div class="mt-4 flex items-start gap-5">
				<div
					class="flex size-20 shrink-0 items-center justify-center rounded-lg bg-primary/15 font-serif text-3xl font-medium text-primary"
					aria-hidden="true"
				>
					{initials}
				</div>
				<div class="min-w-0">
					<h1 class="font-serif text-3xl font-medium tracking-tight sm:text-4xl">{company.name}</h1>
					<p class="mt-1 text-base text-muted-foreground">NEPSE:{company.symbol} Stock Report</p>
					<p class="mt-1 text-sm text-muted-foreground">
						{#if marketCap}<span
								title="Latest close times listed shares; paid-up capital over the Rs 100 par when NEPSE lists no share count"
								>Market Cap Rs {(marketCap / 1e7).toLocaleString('en-NP', {
									maximumFractionDigits: 0
								})} crore</span
							>{/if}
						{#if rating?.letter}
							{#if marketCap}·{/if}
							<span title="A to D from the share of checks that hold; not a recommendation">
								Rating <span class="font-medium text-foreground">{rating.letter}</span>, {rating.holds}
								of {rating.total} checks hold, {rating.evaluated} evaluated
							</span>
						{:else if rating && story}
							{#if marketCap}·{/if}
							Unrated, {rating.evaluated} of {story.checks.length} checks have enough history
						{/if}
					</p>
				</div>
			</div>
			<div class="mt-6 flex flex-wrap items-center gap-2">
				<AIResearchButton {company} {story} />
			</div>
			<p class="mt-6 text-xs text-muted-foreground">
				{#if quote}
					Close of {quote.businessDate}{#if quote.businessDateBs}
						&nbsp;({quote.businessDateBs}){/if} ·
				{/if}
				{story?.lastFiling ? `Last filing ${story.lastFiling}` : 'No filings recorded'} · Data from NEPSE
				closes and company filings
			</p>
		</div>

		{#if quote}
			<div class="flex flex-col lg:pt-1">
				<p class="flex items-center gap-2 text-sm text-muted-foreground">
					<span class="size-2 rounded-full bg-primary" aria-hidden="true"></span> Share price
				</p>
				<p class="mt-1 font-serif text-3xl font-medium tabular-nums">Rs {fmt(quote.close)}</p>
				{#if prices.length > 0}
					<div class="mt-3 h-44">
						<PriceChart {prices} events={story?.events ?? []} days={chartDays} class="h-44" />
					</div>
				{/if}
				<div class="mt-2 flex flex-wrap items-center justify-between gap-3">
					{#if prices.length > 0}
						<TimeRangeSelector selected={chartDays} onSelect={(days) => (chartDays = days)} />
					{/if}
					<p class="flex gap-5 text-sm tabular-nums">
						{#if quote.changePercent !== undefined}
							<span
								><span class="text-muted-foreground">1D</span>
								<span class={tone(quote.changePercent)}>{signed(quote.changePercent)}%</span></span
							>
						{/if}
						{#if yearChange !== undefined}
							<span
								><span class="text-muted-foreground">1Y</span>
								<span class={tone(yearChange)}>{signed(yearChange)}%</span></span
							>
						{/if}
					</p>
				</div>
			</div>
		{/if}
	</div>
</header>
