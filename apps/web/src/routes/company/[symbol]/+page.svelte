<script lang="ts">
	import { generateStory } from '$lib/story';
	import { Sector, CompanyStatus } from '$lib/gen/ntx/v1/common_pb';
	import { PriceChart, EarningsChart, RadarChart } from '$lib/components/charts';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import Sparkles from '@lucide/svelte/icons/sparkles';
	import Copy from '@lucide/svelte/icons/copy';
	import Check from '@lucide/svelte/icons/check';

	let { data } = $props();
	let aiDialogOpen = $state(false);
	let copied = $state(false);
	let company = $derived(data.company);
	let fundamentals = $derived(data.fundamentals);
	let priceData = $derived(data.price);
	let priceHistory = $derived(data.priceHistory);
	let sectorStats = $derived(data.sectorStats);

	let chartDays = $state<30 | 90 | 180 | 365>(365);

	let story = $derived.by(() => {
		if (!company || !priceData || !fundamentals) return null;
		return generateStory({
			company,
			price: priceData,
			priceHistory: priceHistory ?? [],
			fundamentals,
			fundamentalsHistory: data.fundamentalsHistory ?? [],
			sectorStats
		});
	});

	const sectorNames: Record<number, string> = {
		[Sector.COMMERCIAL_BANK]: 'Commercial Banking',
		[Sector.DEVELOPMENT_BANK]: 'Development Banking',
		[Sector.FINANCE]: 'Finance',
		[Sector.MICROFINANCE]: 'Microfinance',
		[Sector.LIFE_INSURANCE]: 'Life Insurance',
		[Sector.NON_LIFE_INSURANCE]: 'Non-Life Insurance',
		[Sector.HYDROPOWER]: 'Hydropower',
		[Sector.MANUFACTURING]: 'Manufacturing',
		[Sector.HOTEL]: 'Hotels & Tourism',
		[Sector.TRADING]: 'Trading',
		[Sector.INVESTMENT]: 'Investment',
		[Sector.MUTUAL_FUND]: 'Mutual Fund',
		[Sector.OTHERS]: 'Others'
	};

	function fmt(value: number | undefined): string {
		if (value === undefined) return '—';
		return value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	function fmtLarge(value: number | undefined): string {
		if (value === undefined) return '—';
		if (value >= 10_000_000) return `${(value / 10_000_000).toFixed(2)} Cr`;
		if (value >= 100_000) return `${(value / 100_000).toFixed(2)} L`;
		return fmt(value);
	}

	const timeRanges = [
		{ days: 30, label: '1M' },
		{ days: 90, label: '3M' },
		{ days: 180, label: '6M' },
		{ days: 365, label: '1Y' }
	] as const;

	// 52W range calculations
	let rangeInfo = $derived.by(() => {
		if (!priceHistory || priceHistory.length === 0) return null;
		const highs = priceHistory.map((p) => p.high ?? p.ltp ?? 0).filter((v) => v > 0);
		const lows = priceHistory.map((p) => p.low ?? p.ltp ?? 0).filter((v) => v > 0);
		if (highs.length === 0 || lows.length === 0) return null;

		const high52w = Math.max(...highs);
		const low52w = Math.min(...lows);
		const current = priceData?.ltp ?? 0;
		const range = high52w - low52w;
		const position = range > 0 ? ((current - low52w) / range) * 100 : 50;
		const fromHigh = high52w > 0 ? ((high52w - current) / high52w) * 100 : 0;
		const fromLow = low52w > 0 ? ((current - low52w) / low52w) * 100 : 0;

		return { high52w, low52w, position, fromHigh, fromLow };
	});

	// Price statistics from history
	let priceStats = $derived.by(() => {
		if (!priceHistory || priceHistory.length === 0) return null;
		const volumes = priceHistory.map((p) => Number(p.volume ?? 0)).filter((v) => v > 0);
		const avgVolume = volumes.length > 0 ? volumes.reduce((a, b) => a + b, 0) / volumes.length : 0;
		const todayVolume = Number(priceData?.volume ?? 0);
		const volumeRatio = avgVolume > 0 ? todayVolume / avgVolume : 1;

		// Calculate recent performance
		const sorted = [...priceHistory].sort((a, b) => b.businessDate.localeCompare(a.businessDate));
		const current = priceData?.ltp ?? 0;
		const week1 = sorted[5]?.ltp ?? sorted[5]?.close ?? current;
		const month1 = sorted[20]?.ltp ?? sorted[20]?.close ?? current;
		const month3 = sorted[60]?.ltp ?? sorted[60]?.close ?? current;

		return {
			avgVolume,
			todayVolume,
			volumeRatio,
			change1W: week1 > 0 ? ((current - week1) / week1) * 100 : 0,
			change1M: month1 > 0 ? ((current - month1) / month1) * 100 : 0,
			change3M: month3 > 0 ? ((current - month3) / month3) * 100 : 0
		};
	});

	// Radar chart data for fundamentals comparison
	let radarData = $derived.by(() => {
		if (!fundamentals) return [];

		const eps = fundamentals.eps ?? 0;
		const pe = fundamentals.peRatio ?? 0;
		const bv = fundamentals.bookValue ?? 0;
		const pbRatio = bv > 0 ? (priceData?.ltp ?? 0) / bv : 0;

		// For P/E, lower is better, so we invert the scale
		const peInverted = pe > 0 ? Math.max(0, 50 - pe) : 0;

		return [
			{
				label: 'EPS',
				value: Math.min(eps, 100),
				max: 100,
				sectorAvg: sectorStats?.avgEps ? Math.min(sectorStats.avgEps, 100) : undefined
			},
			{
				label: 'Book Value',
				value: Math.min(bv, 500),
				max: 500,
				sectorAvg: sectorStats?.avgBookValue ? Math.min(sectorStats.avgBookValue, 500) : undefined
			},
			{
				label: 'P/E (inv)',
				value: peInverted,
				max: 50,
				sectorAvg: sectorStats?.avgPeRatio ? Math.max(0, 50 - sectorStats.avgPeRatio) : undefined
			},
			{
				label: 'P/B Ratio',
				value: Math.min(pbRatio * 10, 50),
				max: 50,
				sectorAvg: undefined
			}
		];
	});

	// Generate opening paragraph
	let openingParagraph = $derived.by(() => {
		if (!company || !priceData || !story) return '';

		const sector = sectorNames[company.sector ?? Sector.OTHERS];
		const price = fmt(priceData.ltp);
		const change = priceData.changePercent ?? 0;
		const changeDir = change > 0 ? 'up' : change < 0 ? 'down' : 'flat';

		let intro = `${company.name} (${company.symbol}) operates in the ${sector} sector. `;
		intro += `The stock currently trades at Rs. ${price}, ${changeDir} ${Math.abs(change).toFixed(2)}% today. `;
		intro += story.price.positionSentence;

		return intro;
	});

	// Generate AI research prompt with all company data
	let aiPrompt = $derived.by(() => {
		if (!company || !priceData) return '';

		const sector = sectorNames[company.sector ?? Sector.OTHERS];
		const today = priceData.businessDate ?? new Date().toISOString().split('T')[0];

		let prompt = `Act as a Senior Financial Analyst specializing in the Nepalese Stock Market (NEPSE).
Your goal is to perform a deep-dive investment analysis of: ${company.name} (${company.symbol}).

## 1. Provided Data Snapshot (As of ${today})
- **Price**: Rs. ${fmt(priceData.ltp)}
- **Sector**: ${sector}
- **Sector Avg P/E**: ${fmt(sectorStats?.avgPeRatio ?? 0)}
- **Fundamentals**:
  - EPS: ${fmt(fundamentals?.eps)}
  - P/E Ratio: ${fmt(fundamentals?.peRatio)}
  - Book Value: ${fmt(fundamentals?.bookValue)}
  - Paid-up Capital: ${fmt(fundamentals?.paidUpCapital)}

## 2. Research Tasks (MANDATORY WEB SEARCH)
Please SEARCH THE WEB (using browsing capabilities) for the following real-time information:
1.  **Recent News**: Look for the latest news on "sharesansar", "merolagani", or "bizmandu" regarding ${company.symbol} in the last 6 months.
2.  **Regulatory Impacts**: Are there any recent NRB directives, BFI regulations, or insurance board policies affecting the ${sector} sector?
3.  **Corporate Actions**: Check for recent AGM announcements, dividend declarations, or right share issues.

## 3. Analysis Requirements
Combine the provided data with your web research to answer:
- **Valuation**: Is ${company.symbol} undervalued compared to its peers in the ${sector} sector? (Compare P/E and P/B).
- **Growth Outlook**: Based on the latest quarterly reports you find, is the company growing its core business?
- **Risk Assessment**: What are the specific regulatory or macro risks for this company right now?

## 4. Investment Verdict
Conclude with a structured verdict:
- **Recommendation**: [Buy / Hold / Sell]
- **Time Horizon**: [Short-term / Long-term]
- **Key Catalyst**: [One specific event to watch]

Please be objective, critical, and data-driven.`;

		return prompt;
	});

	async function copyPrompt() {
		if (!aiPrompt) return;
		await navigator.clipboard.writeText(aiPrompt);
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}
</script>

{#if company && priceData}
	<article class="min-h-screen pb-12">
		<!-- Header -->
		<header class="border-b border-border bg-card">
			<div class="mx-auto max-w-5xl px-4 py-5">
				<a href="/company" class="text-sm text-muted-foreground hover:text-foreground">
					&larr; Companies
				</a>

				<div class="mt-3 flex items-start justify-between gap-6">
					<div>
						<div class="flex items-baseline gap-2">
							<h1 class="text-2xl tracking-tight">{company.symbol}</h1>
							{#if company.status === CompanyStatus.ACTIVE}
								<span
									class="rounded bg-positive/10 px-1.5 py-0.5 text-[10px] font-medium text-positive uppercase"
								>
									Active
								</span>
							{/if}
						</div>
						<p class="text-muted-foreground">{company.name}</p>
						<p class="text-sm text-muted-foreground">
							{sectorNames[company.sector ?? Sector.OTHERS]}
						</p>
					</div>

					<div class="flex items-start gap-4">
						<div class="text-right">
							<p class="text-2xl font-medium tabular-nums">Rs. {fmt(priceData.ltp)}</p>
							<p
								class="tabular-nums {priceData.changePercent && priceData.changePercent > 0
									? 'text-positive'
									: priceData.changePercent && priceData.changePercent < 0
										? 'text-negative'
										: 'text-muted-foreground'}"
							>
								{priceData.change && priceData.change > 0 ? '+' : ''}{fmt(priceData.change)}
								({priceData.changePercent && priceData.changePercent > 0 ? '+' : ''}{fmt(
									priceData.changePercent
								)}%)
							</p>
						</div>
						<Button variant="outline" size="sm" onclick={() => (aiDialogOpen = true)}>
							<Sparkles class="size-4" />
							Use AI
						</Button>
					</div>
				</div>
			</div>
		</header>

		<div class="mx-auto max-w-5xl px-4">
			<!-- Opening Story -->
			{#if story}
				<section class="border-b border-border py-6">
					<p class="text-lg leading-relaxed">{openingParagraph}</p>
				</section>
			{/if}

			<!-- Main Grid: Chart + Stats -->
			<section class="grid gap-6 border-b border-border py-6 lg:grid-cols-3">
				<!-- Price Chart (2 cols) -->
				<div class="lg:col-span-2">
					<div class="mb-3 flex items-center justify-between">
						<h2 class="font-serif text-lg">Price History</h2>
						<div class="flex gap-0.5 rounded bg-muted p-0.5">
							{#each timeRanges as range (range.days)}
								<button
									onclick={() => (chartDays = range.days)}
									class="rounded px-2 py-1 text-xs font-medium {chartDays === range.days
										? 'bg-background shadow-sm'
										: 'text-muted-foreground hover:text-foreground'}"
								>
									{range.label}
								</button>
							{/each}
						</div>
					</div>
					<div class="rounded-lg border border-border bg-card p-3">
						<PriceChart prices={priceHistory ?? []} days={chartDays} />
					</div>
					{#if story}
						<p class="mt-3 text-sm text-muted-foreground">
							{story.price.trendSentence}
							{#if story.price.volumeContext}
								{story.price.volumeContext}
							{/if}
						</p>
					{/if}
				</div>

				<!-- Stats Sidebar -->
				<div class="space-y-4">
					<!-- 52W Range -->
					{#if rangeInfo}
						<div class="rounded-lg border border-border bg-card p-4">
							<h3 class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
								52 Week Range
							</h3>
							<div class="mt-3">
								<div class="flex justify-between text-xs text-muted-foreground">
									<span>Rs. {fmt(rangeInfo.low52w)}</span>
									<span>Rs. {fmt(rangeInfo.high52w)}</span>
								</div>
								<div class="relative mt-1.5 h-1.5 rounded-full bg-muted">
									<div
										class="absolute h-full w-full rounded-full bg-gradient-to-r from-negative via-caution to-positive opacity-30"
									></div>
									<div
										class="absolute top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-background bg-foreground"
										style="left: {rangeInfo.position}%"
									></div>
								</div>
								<div class="mt-2 grid grid-cols-2 gap-2 text-xs">
									<div>
										<span class="text-muted-foreground">From High</span>
										<span class="ml-1 text-negative">-{rangeInfo.fromHigh.toFixed(1)}%</span>
									</div>
									<div class="text-right">
										<span class="text-muted-foreground">From Low</span>
										<span class="ml-1 text-positive">+{rangeInfo.fromLow.toFixed(1)}%</span>
									</div>
								</div>
							</div>
						</div>
					{/if}

					<!-- Performance -->
					{#if priceStats}
						<div class="rounded-lg border border-border bg-card p-4">
							<h3 class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
								Performance
							</h3>
							<div class="mt-3 space-y-2">
								<div class="flex justify-between text-sm">
									<span class="text-muted-foreground">1 Week</span>
									<span
										class="tabular-nums {priceStats.change1W >= 0
											? 'text-positive'
											: 'text-negative'}"
									>
										{priceStats.change1W >= 0 ? '+' : ''}{priceStats.change1W.toFixed(2)}%
									</span>
								</div>
								<div class="flex justify-between text-sm">
									<span class="text-muted-foreground">1 Month</span>
									<span
										class="tabular-nums {priceStats.change1M >= 0
											? 'text-positive'
											: 'text-negative'}"
									>
										{priceStats.change1M >= 0 ? '+' : ''}{priceStats.change1M.toFixed(2)}%
									</span>
								</div>
								<div class="flex justify-between text-sm">
									<span class="text-muted-foreground">3 Months</span>
									<span
										class="tabular-nums {priceStats.change3M >= 0
											? 'text-positive'
											: 'text-negative'}"
									>
										{priceStats.change3M >= 0 ? '+' : ''}{priceStats.change3M.toFixed(2)}%
									</span>
								</div>
							</div>
						</div>
					{/if}

					<!-- Volume -->
					{#if priceStats && priceStats.avgVolume > 0}
						<div class="rounded-lg border border-border bg-card p-4">
							<h3 class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
								Volume
							</h3>
							<div class="mt-3 space-y-2">
								<div class="flex justify-between text-sm">
									<span class="text-muted-foreground">Today</span>
									<span class="tabular-nums">{fmtLarge(priceStats.todayVolume)}</span>
								</div>
								<div class="flex justify-between text-sm">
									<span class="text-muted-foreground">Avg (1Y)</span>
									<span class="tabular-nums">{fmtLarge(priceStats.avgVolume)}</span>
								</div>
								<div class="flex justify-between text-sm">
									<span class="text-muted-foreground">vs Avg</span>
									<span
										class="tabular-nums {priceStats.volumeRatio >= 1
											? 'text-positive'
											: 'text-negative'}"
									>
										{priceStats.volumeRatio.toFixed(2)}x
									</span>
								</div>
							</div>
						</div>
					{/if}
				</div>
			</section>

			<!-- Fundamentals Section -->
			{#if fundamentals && story}
				<section class="border-b border-border py-6">
					<h2 class="font-serif text-xl">Fundamentals</h2>
					<p class="mt-2 text-lg">{story.earnings.headline}</p>
					<p class="mt-1 text-muted-foreground">{story.earnings.detail}</p>

					<!-- Earnings Chart -->
					<div class="mt-6">
						<h3 class="mb-3 text-sm font-medium">Profit History</h3>
						<div class="rounded-lg border border-border bg-card p-3">
							<EarningsChart fundamentals={data.fundamentalsHistory ?? []} />
						</div>
						<p class="mt-2 text-sm text-muted-foreground italic">{story.earnings.trendSentence}</p>
					</div>

					<!-- Radar Chart -->
					<div class="mt-6">
						<h3 class="mb-3 text-sm font-medium">Fundamentals vs Sector</h3>
						<div
							class="flex items-center justify-center rounded-lg border border-border bg-card p-4"
						>
							<RadarChart data={radarData} />
						</div>
					</div>

					<!-- Key Metrics Grid -->
					<div class="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
						<div class="rounded-lg border border-border bg-card p-3">
							<p class="text-xs text-muted-foreground">EPS</p>
							<p class="mt-1 text-xl font-medium tabular-nums">{fmt(fundamentals.eps)}</p>
							{#if sectorStats?.avgEps}
								<p class="text-xs text-muted-foreground">Sector: {fmt(sectorStats.avgEps)}</p>
							{/if}
						</div>
						<div class="rounded-lg border border-border bg-card p-3">
							<p class="text-xs text-muted-foreground">P/E Ratio</p>
							<p class="mt-1 text-xl font-medium tabular-nums">{fmt(fundamentals.peRatio)}</p>
							{#if sectorStats?.avgPeRatio}
								<p class="text-xs text-muted-foreground">Sector: {fmt(sectorStats.avgPeRatio)}</p>
							{/if}
						</div>
						<div class="rounded-lg border border-border bg-card p-3">
							<p class="text-xs text-muted-foreground">Book Value</p>
							<p class="mt-1 text-xl font-medium tabular-nums">{fmt(fundamentals.bookValue)}</p>
							{#if sectorStats?.avgBookValue}
								<p class="text-xs text-muted-foreground">Sector: {fmt(sectorStats.avgBookValue)}</p>
							{/if}
						</div>
						<div class="rounded-lg border border-border bg-card p-3">
							<p class="text-xs text-muted-foreground">P/B Ratio</p>
							<p class="mt-1 text-xl font-medium tabular-nums">
								{fundamentals.bookValue && fundamentals.bookValue > 0
									? ((priceData.ltp ?? 0) / fundamentals.bookValue).toFixed(2)
									: '—'}
							</p>
						</div>
					</div>
				</section>

				<!-- Valuation Analysis -->
				<section class="border-b border-border py-6">
					<h2 class="font-serif text-xl">Valuation</h2>
					<p class="mt-2 text-lg">{story.valuation.headline}</p>
					<p class="mt-1 text-muted-foreground">{story.valuation.peContext}</p>

					<!-- Additional valuation context -->
					<div class="mt-4 rounded-lg bg-muted/50 p-4">
						<p class="text-sm leading-relaxed">
							{#if fundamentals.peRatio && sectorStats?.avgPeRatio}
								{@const peDiff = fundamentals.peRatio - sectorStats.avgPeRatio}
								{#if peDiff < -5}
									At a P/E of {fmt(fundamentals.peRatio)}, {company.symbol} trades at a significant discount
									to the sector average of {fmt(sectorStats.avgPeRatio)}. This could indicate the
									market sees higher risk, or it may represent an undervalued opportunity if
									earnings remain stable.
								{:else if peDiff > 5}
									At a P/E of {fmt(fundamentals.peRatio)}, {company.symbol} trades at a premium to the
									sector average of {fmt(sectorStats.avgPeRatio)}. The market may be pricing in
									expected growth or sees lower risk in this stock compared to peers.
								{:else}
									At a P/E of {fmt(fundamentals.peRatio)}, {company.symbol} trades roughly in line with
									the sector average of {fmt(sectorStats.avgPeRatio)}, suggesting the market views
									it as fairly valued relative to peers.
								{/if}
							{:else if fundamentals.peRatio}
								At a P/E of {fmt(fundamentals.peRatio)}, you're paying Rs. {fmt(
									fundamentals.peRatio
								)} for every rupee of earnings. Compare this with other stocks in the sector to gauge
								relative value.
							{/if}

							{#if fundamentals.bookValue}
								{@const pbRatio = (priceData.ltp ?? 0) / fundamentals.bookValue}
								{#if pbRatio < 1}
									The stock trades below book value (P/B: {pbRatio.toFixed(2)}), meaning you could
									theoretically buy the company for less than its net assets are worth.
								{:else if pbRatio > 3}
									With a P/B ratio of {pbRatio.toFixed(2)}, the market values the company at more
									than
									{pbRatio.toFixed(1)}x its book value, reflecting expectations of strong future
									earnings or intangible value not captured on the balance sheet.
								{/if}
							{/if}
						</p>
					</div>
				</section>

				<!-- The Verdict -->
				<section class="py-6">
					<h2 class="font-serif text-xl">{story.verdict.title}</h2>

					<div
						class="mt-4 rounded-lg p-5 {story.verdict.signal === 'opportunity'
							? 'bg-positive/5 ring-1 ring-positive/20'
							: story.verdict.signal === 'caution'
								? 'bg-caution/5 ring-1 ring-caution/20'
								: 'bg-muted'}"
					>
						<div class="flex items-center gap-2">
							{#if story.verdict.signal === 'opportunity'}
								<div class="h-2 w-2 rounded-full bg-positive"></div>
								<span class="text-sm font-medium text-positive">Potential Opportunity</span>
							{:else if story.verdict.signal === 'caution'}
								<div class="h-2 w-2 rounded-full bg-caution"></div>
								<span class="text-sm font-medium text-caution">Proceed with Caution</span>
							{:else}
								<div class="h-2 w-2 rounded-full bg-muted-foreground"></div>
								<span class="text-sm font-medium text-muted-foreground">Neutral</span>
							{/if}
						</div>
						<p class="mt-3 leading-relaxed">{story.verdict.summary}</p>
					</div>

					<!-- Closing context -->
					<div class="mt-4 text-sm text-muted-foreground">
						<p>
							Remember: No single metric tells the whole story. Consider the company's competitive
							position, management quality, and broader market conditions before making any
							investment decision. Past performance does not guarantee future results.
						</p>
					</div>

					<p class="mt-4 text-xs text-muted-foreground">
						This analysis is generated automatically from available data and should not be
						considered investment advice. Last updated: {priceData.businessDate}
					</p>
				</section>
			{/if}
		</div>

		<!-- Footer -->
		<footer class="border-t border-border">
			<div class="mx-auto flex max-w-5xl items-center justify-between px-4 py-4">
				<a href="/company" class="text-sm text-muted-foreground hover:text-foreground">
					&larr; All companies
				</a>
			</div>
		</footer>
	</article>

	<!-- AI Prompt Dialog -->
	<Dialog.Root bind:open={aiDialogOpen}>
		<Dialog.Content class="max-w-2xl">
			<Dialog.Header>
				<Dialog.Title class="flex items-center gap-2">
					<Sparkles class="size-5" />
					AI Research Prompt
				</Dialog.Title>
				<Dialog.Description>
					Copy this prompt and paste it into your preferred AI assistant (ChatGPT, Claude, etc.) for
					detailed investment research and advice.
				</Dialog.Description>
			</Dialog.Header>
			<div class="relative">
				<pre
					class="max-h-[400px] overflow-auto rounded-lg bg-muted p-4 text-sm whitespace-pre-wrap">{aiPrompt}</pre>
				<Button variant="secondary" size="sm" class="absolute top-2 right-2" onclick={copyPrompt}>
					{#if copied}
						<Check class="size-4" />
						Copied!
					{:else}
						<Copy class="size-4" />
						Copy
					{/if}
				</Button>
			</div>
			<Dialog.Footer>
				<Button variant="outline" onclick={() => (aiDialogOpen = false)}>Close</Button>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{/if}
