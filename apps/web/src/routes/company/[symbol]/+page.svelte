<script lang="ts">
	import {
		StoryOpening,
		ClaimText,
		RatioHistory,
		PeerBars,
		FinancialsTable,
		MutualFundHoldings,
		FiscalYears,
		PayoutBars,
		OwnershipHistory,
		Snowflake
	} from '#lib/components/company/index.js';
	import { checksFor, score, type AxisId } from '#lib/components/company/snowflake.js';
	import { EarningsChart } from '#lib/components/charts/index.js';
	import CompanyPlants from '#lib/components/hydro/CompanyPlants.svelte';
	import Gates from '#lib/components/portfolio/Gates.svelte';
	import { enhance } from '$app/forms';
	import CircleCheck from '@lucide/svelte/icons/circle-check';
	import TriangleAlert from '@lucide/svelte/icons/triangle-alert';
	import CircleDashed from '@lucide/svelte/icons/circle-dashed';
	import type { Claim, Check } from '#lib/gen/ntx/v1/story_pb.js';

	let { data } = $props();

	let company = $derived(data.company);
	let story = $derived(data.story);
	let fundamentalsHistory = $derived(data.fundamentalsHistory ?? []);
	let priceHistory = $derived(data.priceHistory ?? []);
	let fundHoldings = $derived(data.fundHoldings ?? []);
	let plants = $derived(data.plants ?? []);
	let thesis = $derived(data.thesis);
	const field = 'h-8 rounded-md border border-input bg-background px-2 text-sm';
	const bankColumns = [
		['nplPercent', 'NPL %'],
		['carPercent', 'CAR %'],
		['loanGrowthPercent', 'Loan growth %'],
		['distributableProfit', 'Distributable profit Rs']
	] as const;
	const today = new Date().toISOString().slice(0, 10);
	let lastFiling = $derived(
		story?.lastFiling ? `Last filing: ${story.lastFiling}.` : 'No filings recorded.'
	);
	let filedDaysAgo = $derived(
		story?.filedAt ? Math.floor((Date.now() - Date.parse(story.filedAt)) / 86_400_000) : undefined
	);
	let allChecks = $derived(story?.checks ?? []);
	// Listed shares when NEPSE publishes them, else paid-up capital over the Rs 100 par.
	let shares = $derived(
		Number(company?.listedShares ?? 0) || (fundamentalsHistory[0]?.paidUpCapital ?? 0) / 100
	);
	let holds = $derived(allChecks.filter((c) => c.verdict === 'holds'));
	let watch = $derived(allChecks.filter((c) => c.verdict === 'watch'));
	let pending = $derived(allChecks.filter((c) => c.verdict === 'needs more history'));
	// The overview reads the first earnings and valuation sentences; each section repeats its own.
	let lede = $derived(
		['earnings', 'valuation']
			.map((k) => (story?.claims ?? []).find((c) => c.kind === k))
			.filter((c): c is Claim => !!c)
	);

	const strip = (ratio: string) => story?.strips.find((x) => x.ratio === ratio);
	let peCheck = $derived(allChecks.find((c) => c.name === 'PE against its own history'));
	let pbCheck = $derived(allChecks.find((c) => c.name === 'PB against the sector'));
	const RATIO: Record<string, string> = { pe: 'PE', pb: 'PB', eps: 'EPS', yield: 'Dividend yield' };
	function ratioFmt(ratio: string, v: number): string {
		const n = v.toLocaleString('en-NP', { maximumFractionDigits: 2 });
		return ratio === 'eps' ? `Rs ${n}` : ratio === 'yield' ? `${n}%` : `${n}x`;
	}

	function claims(...kinds: string[]): Claim[] {
		return (story?.claims ?? []).filter((c) => kinds.includes(c.kind));
	}

	// A section shows the checks its snowflake axis reads (GLOSSARY: Check).
	function axis(id: AxisId): Check[] {
		return checksFor(allChecks, id);
	}

	const VERDICT: Record<string, string> = {
		holds: 'Holds',
		watch: 'Watch',
		'needs more history': 'Needs more history'
	};
</script>

{#snippet icon(verdict: string)}
	{#if verdict === 'holds'}
		<CircleCheck class="mt-0.5 size-4 shrink-0 text-positive" aria-hidden="true" />
	{:else if verdict === 'watch'}
		<TriangleAlert class="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
	{:else}
		<CircleDashed class="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
	{/if}
{/snippet}

<!-- One chip per section: how many of all its checks hold; one without history counts as not holding. -->
{#snippet chip(verdicts: Check[])}
	{@const s = score(verdicts)}
	<span
		class="inline-flex shrink-0 items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium {s.total ===
			0 || s.evaluated === 0
			? 'border-border text-muted-foreground'
			: s.holds === s.total
				? 'border-positive/40 text-positive'
				: s.holds === 0
					? 'border-negative/40 text-negative'
					: 'border-border text-foreground'}"
		title={verdicts.length
			? verdicts.map((c) => `${c.name}: ${VERDICT[c.verdict]}`).join('; ')
			: 'None of the ten published checks reads this section'}
	>
		{#if s.total === 0}
			<CircleDashed class="size-3.5" aria-hidden="true" /> no checks
		{:else if s.evaluated === 0}
			<CircleDashed class="size-3.5" aria-hidden="true" /> needs more history
		{:else}
			<CircleCheck class="size-3.5" aria-hidden="true" /> {s.holds} of {s.total} hold
		{/if}
	</span>
{/snippet}

<!-- A section opens with its number and name, its chip, the sentence the filings support, then its checks. -->
{#snippet head(n: number, title: string, sentences: Claim[], verdicts: Check[], empty: string)}
	<div class="flex flex-wrap items-center justify-between gap-3">
		<h2 class="font-serif text-2xl font-medium">
			<span class="mr-2 text-muted-foreground">{n}</span>{title}
		</h2>
		{@render chip(verdicts)}
	</div>
	{#if sentences.length > 0 || empty}
		<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">
			{#if sentences.length > 0}
				<ClaimText claims={sentences} />
			{:else}
				<span class="text-muted-foreground">{empty}</span>
			{/if}
		</p>
	{/if}
	{#if verdicts.length > 0}
		<ul class="mt-5 grid gap-x-8 gap-y-3 text-sm sm:grid-cols-2" aria-label="Checks">
			{#each verdicts as check (check.name)}
				<li class="flex gap-2.5">
					{@render icon(check.verdict)}
					<span>
						<span class="font-medium">{check.name}</span>
						<span class="text-muted-foreground"> · {VERDICT[check.verdict]}</span>
						{#if check.verdict !== 'needs more history'}
							<span class="block text-xs text-muted-foreground">{check.rule}</span>
						{/if}
					</span>
				</li>
			{/each}
		</ul>
	{/if}
{/snippet}

{#if company}
	<div class="mx-auto max-w-7xl px-4 py-6">
		<StoryOpening {company} {story} prices={priceHistory} {shares} />

		<!-- Overview: the sentences, the lists, the snowflake -->
		<section id="overview" class="mt-6 card-report">
			<div class="grid gap-8 lg:grid-cols-[1fr_340px]">
				<div class="min-w-0">
					<h2 class="font-serif text-2xl font-medium">{company.symbol} Stock Overview</h2>
					<p class="mt-3 max-w-3xl text-lg leading-relaxed">
						{#if lede.length === 0}
							Not enough filings to say anything yet. {lastFiling}
						{:else}
							<ClaimText claims={lede} />
						{/if}
					</p>

					<h3
						class="mt-7 font-sans text-xs font-semibold tracking-wider text-muted-foreground uppercase"
					>
						Holds
					</h3>
					{#if holds.length > 0}
						<ul class="mt-2 space-y-2">
							{#each holds as check (check.name)}
								<li class="flex gap-2.5">
									{@render icon(check.verdict)}
									<span>{check.rule}</span>
								</li>
							{/each}
						</ul>
					{:else}
						<p class="mt-2 text-sm text-muted-foreground">No check holds yet.</p>
					{/if}

					<h3
						class="mt-6 font-sans text-xs font-semibold tracking-wider text-muted-foreground uppercase"
					>
						Watch
					</h3>
					{#if watch.length > 0}
						<ul class="mt-2 space-y-2">
							{#each watch as check (check.name)}
								<li class="flex gap-2.5">
									{@render icon(check.verdict)}
									<span>{check.rule}</span>
								</li>
							{/each}
						</ul>
					{:else}
						<p class="mt-2 text-sm text-muted-foreground">No check is on watch.</p>
					{/if}

					{#if pending.length > 0}
						<p class="mt-5 text-sm text-muted-foreground">
							{pending.length} of {allChecks.length} checks need more history: {pending
								.map((c) => c.name.toLowerCase())
								.join(', ')}.
						</p>
					{/if}
				</div>

				<div class="flex flex-col items-center lg:items-start">
					<Snowflake checks={allChecks} />
					<p class="mt-2 font-medium">Snowflake</p>
					<p class="mt-1 text-sm text-muted-foreground">
						{#if story?.rating?.letter}
							Rating {story.rating.letter}: {story.rating.holds} of {story.rating.total} checks hold,
							{story.rating.evaluated} evaluated.
						{:else}
							Too few checks have history to rate.
						{/if}
						Each axis is the share of all its checks that hold, from the ten published rules, drawn out
						from a small core. A check on watch or without history scores zero, and no company is rated
						A while a check lacks history. Not a recommendation.
					</p>
				</div>
			</div>
		</section>

		<!-- Your thesis: the owner's own rules for this holding, shown only with the token (ADR 0006) -->
		{#if thesis?.holding}
			{@const h = thesis.holding}
			<section id="thesis" class="mt-6 card-report">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<h2 class="font-serif text-2xl font-medium">Your thesis</h2>
					<a href="/portfolio" class="text-xs text-muted-foreground underline underline-offset-2"
						>Portfolio</a
					>
				</div>
				<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">
					{#if h.thesis}{h.thesis}{:else}<span class="text-muted-foreground"
							>No thesis written. {h.quantity ? `${h.quantity} shares held.` : 'Watched.'}</span
						>{/if}
				</p>
				{#if thesis.bonusHint}
					<p class="mt-2 text-sm text-muted-foreground">{thesis.bonusHint}</p>
				{/if}
				<div class="mt-6 grid gap-6 lg:grid-cols-2">
					<div>
						<Gates symbol={h.symbol} gates={h.gates} />
					</div>
					<div>
						<h3 class="text-xs text-muted-foreground">Notes</h3>
						<ul class="mt-1 space-y-1 text-sm">
							{#each h.notes as note (note.id)}
								<li>
									<span class="text-muted-foreground tabular-nums">{note.day}</span>
									{note.text}
								</li>
							{:else}
								<li class="text-muted-foreground">No notes yet.</li>
							{/each}
						</ul>
						<form
							method="POST"
							action="?/note"
							use:enhance
							class="mt-3 flex flex-wrap gap-2 text-sm"
						>
							<input type="hidden" name="symbol" value={h.symbol} />
							<input type="date" name="day" value={today} required class={field} />
							<input
								type="text"
								name="text"
								required
								aria-label="Note"
								class="{field} min-w-64 flex-1"
							/>
							<button
								type="submit"
								class="h-8 rounded-md border border-input bg-background px-3 text-sm hover:bg-accent"
								>Add</button
							>
						</form>
					</div>
				</div>
				<!-- Bank figures the exchange does not publish; the owner types each with its filing (ADR 0002) -->
				{#if /bank|finance/i.test(h.sector) || thesis.bankMetrics.length}
					<div class="mt-6">
						<h3 class="text-xs text-muted-foreground">Bank figures, as filed</h3>
						<div class="mt-1 overflow-x-auto">
							<table class="w-full text-sm">
								<thead class="text-xs text-muted-foreground">
									<tr>
										<th class="py-1 text-left font-normal">Period</th>
										{#each bankColumns as [, label] (label)}
											<th class="py-1 text-right font-normal">{label}</th>
										{/each}
										<th class="py-1 text-left font-normal">Published</th>
									</tr>
								</thead>
								<tbody>
									{#each thesis.bankMetrics as m (m.fiscalYear + m.quarter)}
										<tr class="border-t border-border/60">
											<td class="py-1">{m.quarter} {m.fiscalYear}</td>
											{#each bankColumns as [key] (key)}
												<td class="py-1 text-right tabular-nums">
													{m[key] === undefined ? 'not recorded' : m[key].toLocaleString('en-IN')}
												</td>
											{/each}
											<td class="py-1">
												<a href={m.sourceUrl} class="underline" rel="noreferrer"
													>{m.publishedDate}</a
												>
											</td>
										</tr>
									{:else}
										<tr
											><td colspan="6" class="py-1 text-muted-foreground">No figures entered.</td
											></tr
										>
									{/each}
								</tbody>
							</table>
						</div>
						<form
							method="POST"
							action="?/bankMetric"
							use:enhance
							class="mt-3 flex flex-wrap items-end gap-2 text-sm"
						>
							<input type="hidden" name="symbol" value={h.symbol} />
							<input
								type="text"
								name="fiscal_year"
								placeholder="2083/84"
								required
								aria-label="Fiscal year"
								class="{field} w-24"
							/>
							<select name="quarter" aria-label="Quarter" class={field}>
								{#each ['Q1', 'Q2', 'Q3', 'Q4'] as q (q)}<option>{q}</option>{/each}
							</select>
							{#each bankColumns as [key, label] (key)}
								<input
									type="number"
									step="any"
									name={key}
									placeholder={label}
									aria-label={label}
									class="{field} w-36"
								/>
							{/each}
							<input
								type="date"
								name="published_date"
								required
								aria-label="Published date"
								class={field}
							/>
							<input
								type="url"
								name="source_url"
								placeholder="https://… filing"
								required
								aria-label="Source URL"
								class="{field} min-w-64 flex-1"
							/>
							<button
								type="submit"
								class="h-8 rounded-md border border-input bg-background px-3 text-sm hover:bg-accent"
								>Save quarter</button
							>
						</form>
					</div>
				{/if}
			</section>
		{/if}

		<!-- 1. Valuation: the four figures, then one chart per check -->
		<section id="valuation" class="mt-6 card-report">
			{@render head(
				1,
				'Valuation',
				[],
				axis('value'),
				story?.strips.length ? '' : `No ratios recorded. ${lastFiling}`
			)}
			{#if story?.strips.length}
				<dl class="mt-8 grid grid-cols-2 gap-x-6 gap-y-4 sm:grid-cols-4">
					{#each story.strips as st (st.ratio)}
						<div>
							<dt class="text-xs text-muted-foreground">{RATIO[st.ratio] ?? st.ratio}</dt>
							<dd class="mt-0.5 font-serif text-2xl font-medium tabular-nums">
								{st.value !== undefined ? ratioFmt(st.ratio, st.value) : '—'}
							</dd>
							<dd class="text-xs text-muted-foreground tabular-nums">
								{#if st.sector}
									sector median {ratioFmt(st.ratio, st.sector.median)}
								{:else}
									no sector figures
								{/if}
							</dd>
						</div>
					{/each}
				</dl>
				<p class="mt-2 text-xs text-muted-foreground">
					From {story.lastFiling || 'the latest filing'} and the latest close; PE is on trailing four-quarter
					EPS where four consecutive quarters exist, EPS is annualized as filed. Sector medians are over
					each sector company's latest filed quarter.
				</p>
			{/if}

			<div class="mt-10 grid gap-10 lg:grid-cols-2">
				<!-- Each chart answers one check; the API's verdict sits beside the question. -->
				<div id="strip-pe" class="min-w-0 scroll-mt-6">
					<h3 class="font-serif text-lg font-medium">Is the PE below its own three-year median?</h3>
					{#if peCheck}
						<p class="mt-1 flex items-center gap-2 text-sm text-muted-foreground">
							{@render icon(peCheck.verdict)}{VERDICT[peCheck.verdict]}
						</p>
					{/if}
					<RatioHistory
						quarters={story?.quarters ?? []}
						median={strip('pe')?.own?.median}
						current={strip('pe')?.value}
						class="mt-6"
					/>
				</div>
				<div id="strip-pb" class="min-w-0 scroll-mt-6">
					<h3 class="font-serif text-lg font-medium">Is the PB below the sector median?</h3>
					{#if pbCheck}
						<p class="mt-1 flex items-center gap-2 text-sm text-muted-foreground">
							{@render icon(pbCheck.verdict)}{VERDICT[pbCheck.verdict]}
						</p>
					{/if}
					<PeerBars
						peers={story?.peers ?? []}
						metric="pb"
						symbol={company.symbol}
						median={strip('pb')?.sector?.median}
						period={story?.lastFiling ?? ''}
						class="mt-6"
					/>
				</div>
			</div>
		</section>

		<!-- 2. Past performance: quarter-only bars, hydro EPS line, rows as filed -->
		<section id="earnings" class="mt-6 card-report">
			{@render head(
				2,
				'Past Performance',
				claims('earnings', 'record'),
				axis('past'),
				`No quarterly filings. ${lastFiling}`
			)}
			{#if story?.hydroEps}
				{@const h = story.hydroEps}
				<p class="mt-6 text-xs text-muted-foreground tabular-nums">
					EPS Rs {h.quarter.toFixed(2)} quarter · Rs {h.annualized.toFixed(2)} annualized · Rs
					{h.fullYear.toFixed(2)}
					{h.fullYearPeriod}
				</p>
			{/if}
			{#if story?.quarters.length}
				<div class="mt-6">
					<EarningsChart quarters={story.quarters} />
				</div>
			{/if}
			{#if fundamentalsHistory.some((f) => !!f.quarter)}
				<FinancialsTable fundamentals={fundamentalsHistory} class="mt-6" />
			{/if}
		</section>

		<!-- 3. Financial health: book value and the fiscal-year rows with compounded growth -->
		<section id="health" class="mt-6 card-report">
			{@render head(
				3,
				'Financial Health',
				[],
				axis('health'),
				story?.fiscalYears.length
					? 'Book value, profit and shares per fiscal year, from full-year filings.'
					: `No full-year filings. ${lastFiling}`
			)}
			{#if story?.fiscalYears.length}
				<div class="mt-6">
					<FiscalYears years={story.fiscalYears} growth={story.growth} />
				</div>
			{/if}
		</section>

		<!-- 4. Dividend: cash and bonus per fiscal year -->
		<section id="payout" class="mt-6 card-report">
			{@render head(
				4,
				'Dividend',
				claims('payout'),
				axis('dividend'),
				story?.fiscalYears.length
					? 'No dividend notice stored yet, so the payout is not recorded.'
					: `No full-year filings or notices. ${lastFiling}`
			)}
			<!-- Bars only once a notice has arrived; a chart of gaps says nothing. -->
			{#if story?.fiscalYears.some((y) => y.payout !== 'not recorded')}
				<PayoutBars years={story.fiscalYears} class="mt-6" />
			{/if}
		</section>

		<!-- 5. Ownership over time -->
		<section id="ownership" class="mt-6 card-report">
			{@render head(
				5,
				'Ownership',
				claims('ownership'),
				axis('ownership'),
				'No shareholding observations recorded.'
			)}
			{#if story?.ownership.length}
				<div class="mt-6">
					<OwnershipHistory observations={story.ownership} />
				</div>
			{/if}
		</section>

		<!-- 6. Fund holdings -->
		<section id="funds" class="mt-6 card-report">
			{@render head(
				6,
				'Funds',
				claims('funds'),
				[],
				`No open-ended fund lists it in the NAV reports for ${data.navMonths || 'no month on file'}.`
			)}
			{#if fundHoldings.length > 0}
				<div class="mt-6">
					<MutualFundHoldings holdings={fundHoldings} />
				</div>
			{/if}
		</section>

		<!-- 7. Plants, hydro companies only: river, basin, MW and today's state -->
		{#if plants.length > 0}
			<section id="plants" class="mt-6 card-report">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<h2 class="font-serif text-2xl font-medium">
						<span class="mr-2 text-muted-foreground">7</span>Plants
					</h2>
					{@render chip([])}
				</div>
				<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">{data.plantsOpening}</p>
				<div class="mt-6">
					<CompanyPlants {plants} />
				</div>
			</section>
		{/if}

		<!-- How fresh the data is -->
		<footer class="mt-8 px-2 text-xs text-muted-foreground">
			<p class="tabular-nums">
				Last sync {story?.syncedAt ? `${story.syncedAt} UTC` : 'not recorded'} · Newest filing
				{#if story?.filedAt}
					{story.lastFiling}, {story.filedAtSeen ? 'first seen by NTX on' : 'published'}
					{story.filedAt} ({filedDaysAgo} days ago)
				{:else}
					none
				{/if}
				·
				<a href="#earnings" class="underline hover:text-foreground">raw rows</a>
			</p>
		</footer>
	</div>
{/if}
