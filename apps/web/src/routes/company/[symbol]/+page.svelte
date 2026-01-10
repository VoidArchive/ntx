<script lang="ts">
	import { Sector, CompanyStatus } from '$lib/gen/ntx/v1/common_pb';
	import { PriceChart } from '$lib/components/charts';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import Sparkles from '@lucide/svelte/icons/sparkles';
	import Copy from '@lucide/svelte/icons/copy';
	import Check from '@lucide/svelte/icons/check';
	import ExternalLink from '@lucide/svelte/icons/external-link';

	let { data } = $props();
	let aiDialogOpen = $state(false);
	let copied = $state(false);
	let company = $derived(data.company);
	let fundamentals = $derived(data.fundamentals);
	let priceData = $derived(data.price);
	let priceHistory = $derived(data.priceHistory);
	let sectorStats = $derived(data.sectorStats);

	let currentPrice = $derived(priceData?.ltp ?? priceData?.close);

	let chartDays = $state<number>(365);

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

	function fmt(value: number | bigint | undefined): string {
		if (value === undefined) return '—';
		return Number(value).toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	function fmtLarge(value: number | bigint | undefined): string {
		if (value === undefined) return '—';
		const num = Number(value);
		if (num >= 10_000_000) return `${(num / 10_000_000).toFixed(2)} Cr`;
		if (num >= 100_000) return `${(num / 100_000).toFixed(2)} L`;
		return fmt(num);
	}

	const timeRanges = [
		{ days: 7, label: '1W' },
		{ days: 30, label: '1M' },
		{ days: 90, label: '3M' },
		{ days: 180, label: '6M' },
		{ days: 365, label: '1Y' },
		{ days: 0, label: 'All' }
	];

	// 52W range calculations
	let rangeInfo = $derived.by(() => {
		if (!priceHistory || priceHistory.length === 0) return null;
		const highs = priceHistory.map((p) => p.high ?? p.ltp ?? 0).filter((v) => v > 0);
		const lows = priceHistory.map((p) => p.low ?? p.ltp ?? 0).filter((v) => v > 0);
		if (highs.length === 0 || lows.length === 0) return null;

		const high52w = Math.max(...highs);
		const low52w = Math.min(...lows);

		return { high52w, low52w };
	});

	// Generate AI research prompt
	let aiPrompt = $derived.by(() => {
		if (!company || !priceData) return '';

		const sector = sectorNames[company.sector ?? Sector.OTHERS];
		const today = priceData.businessDate ?? new Date().toISOString().split('T')[0];

		return `Act as a Senior Financial Analyst specializing in the Nepalese Stock Market (NEPSE).
Your goal is to perform a deep-dive investment analysis of: ${company.name} (${company.symbol}).

## 1. Provided Data Snapshot (As of ${today})
- **Price**: Rs. ${fmt(currentPrice)}
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
	});

	async function copyPrompt() {
		if (!aiPrompt) return;
		await navigator.clipboard.writeText(aiPrompt);
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}

	function getFilteredDays(days: number): number {
		if (days === 0) return priceHistory?.length ?? 365;
		return days;
	}
</script>

{#if company && priceData}
	<article class="min-h-screen">
		<!-- Compact Header -->
		<header class="border-b border-border bg-card">
			<div class="mx-auto max-w-6xl px-4 py-4">
				<div class="flex flex-wrap items-center justify-between gap-4">
					<div class="flex flex-wrap items-center gap-x-3 gap-y-1">
						<h1 class="text-xl font-semibold tracking-tight">{company.symbol}</h1>
						<span class="text-muted-foreground">{company.name}</span>
						<span class="hidden text-muted-foreground sm:inline">•</span>
						<span class="hidden text-muted-foreground sm:inline">NEPSE</span>
						<span class="hidden text-muted-foreground sm:inline">•</span>
						<span class="text-xl tabular-nums">Rs. {fmt(currentPrice)}</span>
						<span
							class="tabular-nums {priceData.changePercent && priceData.changePercent > 0
								? 'text-positive'
								: priceData.changePercent && priceData.changePercent < 0
									? 'text-negative'
									: 'text-muted-foreground'}"
						>
							{#if priceData.changePercent && priceData.changePercent < 0}▼{:else if priceData.changePercent && priceData.changePercent > 0}▲{/if}
							{priceData.changePercent && priceData.changePercent > 0 ? '+' : ''}{fmt(
								priceData.changePercent
							)}%
							<span class="text-sm">({priceData.change && priceData.change > 0 ? '+' : ''}{fmt(priceData.change)})</span>
						</span>
					</div>
					<Button variant="outline" size="sm" onclick={() => (aiDialogOpen = true)}>
						<Sparkles class="size-4" />
						<span class="hidden sm:inline">Use AI</span>
					</Button>
				</div>
			</div>
		</header>

		<!-- Time Range Selector -->
		<div class="border-b border-border bg-background">
			<div class="mx-auto max-w-6xl px-4 py-3">
				<div class="flex gap-1">
					{#each timeRanges as range (range.label)}
						<button
							onclick={() => (chartDays = range.days === 0 ? (priceHistory?.length ?? 365) : range.days)}
							class="rounded px-3 py-1.5 text-sm font-medium transition-colors {chartDays === (range.days === 0 ? (priceHistory?.length ?? 365) : range.days)
								? 'bg-foreground text-background'
								: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
						>
							{range.label}
						</button>
					{/each}
				</div>
			</div>
		</div>

		<!-- Main Content: 3-column layout -->
		<div class="mx-auto max-w-6xl px-4 py-6">
			<div class="grid gap-6 lg:grid-cols-[1fr_180px_220px]">
				<!-- Price Chart -->
				<div class="min-w-0">
					<PriceChart prices={priceHistory ?? []} days={chartDays} />
				</div>

				<!-- Key Stats -->
				<div class="space-y-0 text-sm">
					{#if rangeInfo}
						<div class="flex justify-between border-b border-border py-2">
							<span class="text-muted-foreground">52w High</span>
							<span class="tabular-nums font-medium">Rs. {fmt(rangeInfo.high52w)}</span>
						</div>
						<div class="flex justify-between border-b border-border py-2">
							<span class="text-muted-foreground">52w Low</span>
							<span class="tabular-nums font-medium">Rs. {fmt(rangeInfo.low52w)}</span>
						</div>
					{/if}
					{#if fundamentals?.peRatio}
						<div class="flex justify-between border-b border-border py-2">
							<span class="text-muted-foreground">P/E</span>
							<span class="tabular-nums font-medium">{fmt(fundamentals.peRatio)}</span>
						</div>
					{/if}
					{#if fundamentals?.eps}
						<div class="flex justify-between border-b border-border py-2">
							<span class="text-muted-foreground">EPS</span>
							<span class="tabular-nums font-medium">{fmt(fundamentals.eps)}</span>
						</div>
					{/if}
					{#if fundamentals?.bookValue}
						<div class="flex justify-between border-b border-border py-2">
							<span class="text-muted-foreground">Book Value</span>
							<span class="tabular-nums font-medium">{fmt(fundamentals.bookValue)}</span>
						</div>
					{/if}
					<div class="flex justify-between border-b border-border py-2">
						<span class="text-muted-foreground">Volume</span>
						<span class="tabular-nums font-medium">{fmtLarge(priceData.volume)}</span>
					</div>
					<div class="flex justify-between py-2">
						<span class="text-muted-foreground">Turnover</span>
						<span class="tabular-nums font-medium">{fmtLarge(priceData.turnover)}</span>
					</div>
				</div>

				<!-- About Section -->
				<div class="text-sm">
					<h3 class="mb-3 font-medium">About {company.name}</h3>
					{#if company.website}
						<a
							href={company.website.startsWith('http') ? company.website : `https://${company.website}`}
							target="_blank"
							rel="noopener noreferrer"
							class="inline-flex items-center gap-1 text-chart-1 hover:underline"
						>
							{company.website}
							<ExternalLink class="size-3" />
						</a>
					{/if}
					<p class="mt-3 leading-relaxed text-muted-foreground">
						{company.name} operates in the {sectorNames[company.sector ?? Sector.OTHERS]} sector on the
						Nepal Stock Exchange.
					</p>
					{#if company.status === CompanyStatus.ACTIVE}
						<p class="mt-2 text-xs text-positive">Currently trading</p>
					{:else if company.status === CompanyStatus.SUSPENDED}
						<p class="mt-2 text-xs text-caution">Trading suspended</p>
					{/if}
				</div>
			</div>
		</div>
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
					detailed investment research.
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
