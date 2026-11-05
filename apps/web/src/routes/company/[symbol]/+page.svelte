<script lang="ts">
	import {
		StoryOpening,
		ClaimText,
		ValuationStrips,
		TimeRangeSelector,
		FinancialsTable,
		AIResearchButton,
		MutualFundHoldings,
		FiscalYears,
		PayoutBars,
		OwnershipHistory,
		Peers
	} from '#lib/components/company/index.js';
	import { PriceChart, VolumeChart, EarningsChart } from '#lib/components/charts/index.js';
	import type { Claim, Check } from '#lib/gen/ntx/v1/story_pb.js';

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

	function claims(...kinds: string[]): Claim[] {
		return (story?.claims ?? []).filter((c) => kinds.includes(c.kind));
	}

	// Each section shows the checks that read its own figures (GLOSSARY: Check).
	function checks(...names: string[]): Check[] {
		return (story?.checks ?? []).filter((c) => names.includes(c.name));
	}

	const GLYPH: Record<string, string> = { holds: '✓', watch: '△', 'needs more history': '·' };
</script>

<!-- A section opens with its claim as the headline and its checks beside it. -->
{#snippet head(label: string, sentences: Claim[], verdicts: Check[], empty: string)}
	<div class="mb-6 flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
		<div class="max-w-3xl min-w-0 flex-1">
			<h2 class="text-xs tracking-wide text-muted-foreground uppercase">{label}</h2>
			<p class="mt-1 font-serif text-xl leading-snug font-medium">
				{#if sentences.length > 0}
					<ClaimText claims={sentences} />
				{:else}
					<span class="font-normal text-muted-foreground">{empty}</span>
				{/if}
			</p>
		</div>
		{#if verdicts.length > 0}
			<ul class="shrink-0 space-y-1 text-sm lg:w-72" aria-label="Checks">
				{#each verdicts as check (check.name)}
					<li class={check.verdict === 'holds' ? '' : 'text-muted-foreground'}>
						<span aria-hidden="true">{GLYPH[check.verdict]}</span>
						<span class="font-medium">{check.name}</span>
						<span class="text-muted-foreground">
							· {check.verdict === 'needs more history' ? 'needs more history' : check.rule}
						</span>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
{/snippet}

{#if company}
	<div class="min-h-screen bg-background text-foreground">
		<div class="mx-auto max-w-7xl px-4 py-8">
			<!-- 1. Lede: symbol, rating, the earnings sentence, the quote -->
			<StoryOpening {company} {story} />

			<!-- 2. Earnings: quarter-only bars, hydro EPS line, rows as filed -->
			<section id="earnings" class="mt-12 border-t border-border pt-8">
				{@render head(
					'Earnings',
					claims('earnings', 'record'),
					checks('Quarterly profit', 'EPS growth'),
					`No quarterly filings. ${lastFiling}`
				)}
				{#if story?.hydroEps}
					{@const h = story.hydroEps}
					<p class="mb-3 text-xs text-muted-foreground tabular-nums">
						EPS Rs {h.quarter.toFixed(2)} quarter · Rs {h.annualized.toFixed(2)} annualized · Rs
						{h.fullYear.toFixed(2)}
						{h.fullYearPeriod}
					</p>
				{/if}
				{#if story?.quarters.length}
					<EarningsChart quarters={story.quarters} />
				{/if}
				{#if fundamentalsHistory.some((f) => !!f.quarter)}
					<FinancialsTable fundamentals={fundamentalsHistory} class="mt-6" />
				{/if}
			</section>

			<!-- 3. Valuation: four strips, peers beneath on the PE strip's range -->
			<section id="valuation" class="mt-12 border-t border-border pt-8">
				{@render head(
					'Valuation',
					claims('valuation'),
					checks('PE against its own history', 'PB against the sector'),
					`No ratios recorded. ${lastFiling}`
				)}
				{#if story?.strips.length}
					<ValuationStrips strips={story.strips} period={story.lastFiling} />
				{/if}
				{#if story?.peers.length}
					<Peers
						peers={story.peers}
						range={story.strips.find((s) => s.ratio === 'pe')?.sector}
						symbol={company.symbol}
						period={story.lastFiling}
					/>
				{/if}
			</section>

			<!-- 4. Payout: cash and bonus per fiscal year, the year rows, compounded growth -->
			<section id="payout" class="mt-12 border-t border-border pt-8">
				{@render head(
					'Payout',
					claims('payout'),
					checks('Dividend record', 'Cash payout', 'Book value growth'),
					story?.fiscalYears.length
						? 'No dividend notice stored yet, so the payout is not recorded; the fiscal years below are from filings.'
						: `No full-year filings or notices. ${lastFiling}`
				)}
				{#if story?.fiscalYears.length}
					<!-- Bars only once a notice has arrived; a chart of gaps says nothing. -->
					{#if story.fiscalYears.some((y) => y.payout !== 'not recorded')}
						<PayoutBars years={story.fiscalYears} class="mb-8" />
					{/if}
					<FiscalYears years={story.fiscalYears} growth={story.growth} />
				{/if}
			</section>

			<!-- 5. Ownership over time -->
			<section id="ownership" class="mt-12 border-t border-border pt-8">
				{@render head(
					'Ownership',
					claims('ownership'),
					checks('Promoter holding'),
					'No shareholding observations recorded.'
				)}
				{#if story?.ownership.length}
					<OwnershipHistory observations={story.ownership} />
				{/if}
			</section>

			<!-- 6. Fund holdings -->
			<section id="funds" class="mt-12 border-t border-border pt-8">
				{@render head(
					'Funds',
					claims('funds'),
					[],
					`No open-ended fund lists it in the NAV reports for ${data.navMonths || 'no month on file'}.`
				)}
				{#if fundHoldings.length > 0}
					<MutualFundHoldings holdings={fundHoldings} />
				{/if}
			</section>

			<!-- 7. Price and volume last: the quote is context, not the hero -->
			<section id="price" class="mt-12 border-t border-border pt-8">
				<div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
					<h2 class="text-xs tracking-wide text-muted-foreground uppercase">Price and volume</h2>
					{#if priceHistory.length > 0}
						<TimeRangeSelector selected={chartDays} onSelect={(days) => (chartDays = days)} />
					{/if}
				</div>
				{#if priceHistory.length > 0}
					<div class="h-[350px] overflow-hidden">
						<PriceChart prices={priceHistory} events={story?.events ?? []} days={chartDays} />
					</div>
					<div class="mt-2">
						<VolumeChart prices={priceHistory} days={chartDays} />
					</div>
				{:else}
					<p class="text-sm text-muted-foreground">No closes recorded. {lastFiling}</p>
				{/if}
			</section>

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
					<a href="#earnings" class="underline underline-offset-2 hover:text-foreground">raw rows</a
					>
				</p>
				<AIResearchButton {company} {story} />
			</footer>
		</div>
	</div>
{/if}
