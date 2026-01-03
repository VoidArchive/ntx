<script lang="ts">
	import { goto } from '$app/navigation';
	import { PageContainer } from '$lib/components/layout';
	import { Badge } from '$lib/components/ui/badge';
	import * as Table from '$lib/components/ui/table';
	import { getSectorName, getSectorColor } from '$lib/utils/sector';
	import { formatVolume, formatPriceCompact, formatDate, formatNumber, formatPercent, formatMarketCap } from '$lib/utils/format';
	import {
		TrendingUpIcon,
		TrendingDownIcon,
		MinusIcon,
		ArrowLeftIcon,
		CalendarIcon,
		BarChart3Icon,
		UsersIcon
	} from '@lucide/svelte';
	import { ReportType } from '@ntx/api/ntx/v1/common_pb';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const { company, price: stockPrice, fundamentals, reports, candles, sectorPeers } = data;

	// Calculate price position in 52-week range
	const range52w = $derived(stockPrice ? stockPrice.week52High - stockPrice.week52Low : 0);
	const position52w = $derived(
		range52w > 0 && stockPrice
			? ((stockPrice.ltp - stockPrice.week52Low) / range52w) * 100
			: 50
	);

	// Price change indicators
	const isPositive = $derived(stockPrice && stockPrice.change > 0);
	const isNegative = $derived(stockPrice && stockPrice.change < 0);

	// Format report period
	function formatPeriod(report: typeof reports[0]): string {
		if (report.type === ReportType.QUARTERLY) {
			return `Q${report.quarter} FY${report.fiscalYear}`;
		}
		return `FY ${report.fiscalYear}`;
	}
</script>

<svelte:head>
	<title>{company.symbol} - {company.name} | NTX</title>
	<meta
		name="description"
		content="View {company.name} ({company.symbol}) stock price, fundamentals, and financial data."
	/>
</svelte:head>

<!-- Sticky Header -->
<div class="sticky top-14 z-40 border-b bg-background/95 backdrop-blur">
	<PageContainer>
		<div class="flex items-center justify-between py-3">
			<div class="flex items-center gap-3">
				<button onclick={() => history.back()} class="rounded-md p-1 hover:bg-accent">
					<ArrowLeftIcon class="h-5 w-5" />
				</button>
				<div class="flex items-center gap-2">
					<div class="flex h-8 w-8 items-center justify-center rounded-md bg-primary/10 text-sm font-bold text-primary">
						{company.symbol.slice(0, 2)}
					</div>
					<div>
						<span class="font-mono text-lg font-bold">{company.symbol}</span>
						<Badge variant="outline" class="ml-2 {getSectorColor(company.sector)}">
							{getSectorName(company.sector)}
						</Badge>
					</div>
				</div>
			</div>
			{#if stockPrice}
				<div class="text-right">
					<div class="font-mono text-xl font-bold tabular-nums">
						Rs. {formatPriceCompact(stockPrice.ltp)}
					</div>
					<div class="flex items-center justify-end gap-1 text-sm {isPositive ? 'text-positive' : isNegative ? 'text-negative' : 'text-muted-foreground'}">
						{#if isPositive}
							<TrendingUpIcon class="h-4 w-4" />
						{:else if isNegative}
							<TrendingDownIcon class="h-4 w-4" />
						{:else}
							<MinusIcon class="h-4 w-4" />
						{/if}
						<span class="font-mono tabular-nums">
							{stockPrice.change >= 0 ? '+' : ''}{stockPrice.change.toFixed(2)} ({stockPrice.percentChange >= 0 ? '+' : ''}{stockPrice.percentChange.toFixed(2)}%)
						</span>
					</div>
				</div>
			{/if}
		</div>
	</PageContainer>
</div>

<section class="py-6">
	<PageContainer>
		<div class="grid gap-6 lg:grid-cols-3">
			<!-- Main Content -->
			<div class="lg:col-span-2 space-y-6">
				<!-- Company Name -->
				<div>
					<h1 class="text-xl font-semibold text-muted-foreground">{company.name}</h1>
					{#if company.description}
						<p class="mt-2 text-sm text-muted-foreground">{company.description}</p>
					{/if}
				</div>

				<!-- Price Overview Card -->
				{#if stockPrice}
					<div class="rounded-xl border bg-card p-6">
						<h2 class="text-sm font-medium text-muted-foreground mb-4">Price Overview</h2>
						<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-6">
							<div>
								<div class="text-xs text-muted-foreground">Open</div>
								<div class="font-mono font-medium tabular-nums">{formatPriceCompact(stockPrice.open)}</div>
							</div>
							<div>
								<div class="text-xs text-muted-foreground">High</div>
								<div class="font-mono font-medium tabular-nums text-positive">{formatPriceCompact(stockPrice.high)}</div>
							</div>
							<div>
								<div class="text-xs text-muted-foreground">Low</div>
								<div class="font-mono font-medium tabular-nums text-negative">{formatPriceCompact(stockPrice.low)}</div>
							</div>
							<div>
								<div class="text-xs text-muted-foreground">Prev. Close</div>
								<div class="font-mono font-medium tabular-nums">{formatPriceCompact(stockPrice.previousClose)}</div>
							</div>
							<div>
								<div class="text-xs text-muted-foreground">Volume</div>
								<div class="font-mono font-medium tabular-nums">{formatVolume(stockPrice.volume)}</div>
							</div>
							<div>
								<div class="text-xs text-muted-foreground">Turnover</div>
								<div class="font-mono font-medium tabular-nums">{formatVolume(stockPrice.turnover)}</div>
							</div>
						</div>
					</div>
				{/if}

				<!-- 52-Week Range -->
				{#if stockPrice && (stockPrice.week52High > 0 || stockPrice.week52Low > 0)}
					<div class="rounded-xl border bg-card p-6">
						<h2 class="text-sm font-medium text-muted-foreground mb-4">52-Week Range</h2>
						<div class="flex justify-between text-sm mb-2">
							<span class="font-mono tabular-nums text-negative">{formatPriceCompact(stockPrice.week52Low)}</span>
							<span class="font-mono tabular-nums text-positive">{formatPriceCompact(stockPrice.week52High)}</span>
						</div>
						<div class="relative h-3 rounded-full bg-muted">
							<div
								class="absolute top-0 h-3 rounded-full bg-gradient-to-r from-negative via-muted-foreground to-positive opacity-30"
								style="width: 100%"
							></div>
							<div
								class="absolute top-1/2 h-5 w-5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-primary bg-background shadow-sm"
								style="left: {position52w}%"
							></div>
						</div>
						<div class="mt-2 text-center">
							<span class="text-sm text-muted-foreground">Current: </span>
							<span class="font-mono font-medium tabular-nums">Rs. {formatPriceCompact(stockPrice.ltp)}</span>
						</div>
					</div>
				{/if}

				<!-- Financial Reports -->
				{#if reports.length > 0}
					<div class="rounded-xl border bg-card">
						<div class="flex items-center justify-between border-b px-6 py-4">
							<div class="flex items-center gap-2">
								<CalendarIcon class="h-4 w-4 text-muted-foreground" />
								<h2 class="font-semibold">Financial Reports</h2>
							</div>
						</div>
						<div class="overflow-x-auto">
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>Period</Table.Head>
										<Table.Head class="text-right">Revenue</Table.Head>
										<Table.Head class="text-right">Net Income</Table.Head>
										<Table.Head class="text-right">EPS</Table.Head>
										<Table.Head class="text-right">Book Value</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each reports.slice(0, 8) as report}
										<Table.Row>
											<Table.Cell class="font-medium">{formatPeriod(report)}</Table.Cell>
											<Table.Cell class="text-right font-mono tabular-nums">
												{report.revenue ? `Rs. ${formatNumber(report.revenue / 1_000_000, 1)}M` : '-'}
											</Table.Cell>
											<Table.Cell class="text-right font-mono tabular-nums">
												{report.netIncome ? `Rs. ${formatNumber(report.netIncome / 1_000_000, 1)}M` : '-'}
											</Table.Cell>
											<Table.Cell class="text-right font-mono tabular-nums">
												{report.eps ? formatNumber(report.eps, 2) : '-'}
											</Table.Cell>
											<Table.Cell class="text-right font-mono tabular-nums">
												{report.bookValue ? formatNumber(report.bookValue, 2) : '-'}
											</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</div>
					</div>
				{/if}

				<!-- Price History -->
				{#if candles.length > 0}
					<div class="rounded-xl border bg-card">
						<div class="flex items-center justify-between border-b px-6 py-4">
							<div class="flex items-center gap-2">
								<BarChart3Icon class="h-4 w-4 text-muted-foreground" />
								<h2 class="font-semibold">Price History</h2>
							</div>
							<span class="text-xs text-muted-foreground">Last 30 days</span>
						</div>
						<div class="overflow-x-auto max-h-[400px]">
							<Table.Root>
								<Table.Header class="sticky top-0 bg-card">
									<Table.Row>
										<Table.Head>Date</Table.Head>
										<Table.Head class="text-right">Open</Table.Head>
										<Table.Head class="text-right">High</Table.Head>
										<Table.Head class="text-right">Low</Table.Head>
										<Table.Head class="text-right">Close</Table.Head>
										<Table.Head class="text-right">Volume</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each candles.slice().reverse() as candle}
										<Table.Row>
											<Table.Cell class="font-medium">
												{candle.date ? formatDate(new Date(Number(candle.date.seconds) * 1000)) : '-'}
											</Table.Cell>
											<Table.Cell class="text-right font-mono tabular-nums">{formatPriceCompact(candle.open)}</Table.Cell>
											<Table.Cell class="text-right font-mono tabular-nums text-positive">{formatPriceCompact(candle.high)}</Table.Cell>
											<Table.Cell class="text-right font-mono tabular-nums text-negative">{formatPriceCompact(candle.low)}</Table.Cell>
											<Table.Cell class="text-right font-mono tabular-nums">{formatPriceCompact(candle.close)}</Table.Cell>
											<Table.Cell class="text-right font-mono tabular-nums">{formatVolume(candle.volume)}</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</div>
					</div>
				{/if}
			</div>

			<!-- Sidebar -->
			<div class="space-y-6">
				<!-- Key Statistics -->
				{#if fundamentals}
					<div class="rounded-xl border bg-card p-6">
						<h2 class="font-semibold mb-4">Key Statistics</h2>
						<dl class="space-y-3">
							<div class="flex justify-between">
								<dt class="text-sm text-muted-foreground">Market Cap</dt>
								<dd class="font-mono font-medium tabular-nums">
									{fundamentals.marketCap ? formatMarketCap(fundamentals.marketCap) : '-'}
								</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-sm text-muted-foreground">P/E Ratio</dt>
								<dd class="font-mono font-medium tabular-nums">
									{fundamentals.pe ? formatNumber(fundamentals.pe, 2) : '-'}
								</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-sm text-muted-foreground">P/B Ratio</dt>
								<dd class="font-mono font-medium tabular-nums">
									{fundamentals.pb ? formatNumber(fundamentals.pb, 2) : '-'}
								</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-sm text-muted-foreground">EPS</dt>
								<dd class="font-mono font-medium tabular-nums">
									{fundamentals.eps ? `Rs. ${formatNumber(fundamentals.eps, 2)}` : '-'}
								</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-sm text-muted-foreground">Book Value</dt>
								<dd class="font-mono font-medium tabular-nums">
									{fundamentals.bookValue ? `Rs. ${formatNumber(fundamentals.bookValue, 2)}` : '-'}
								</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-sm text-muted-foreground">ROE</dt>
								<dd class="font-mono font-medium tabular-nums">
									{fundamentals.roe ? formatPercent(fundamentals.roe) : '-'}
								</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-sm text-muted-foreground">Dividend Yield</dt>
								<dd class="font-mono font-medium tabular-nums">
									{fundamentals.dividendYield ? formatPercent(fundamentals.dividendYield) : '-'}
								</dd>
							</div>
							<div class="flex justify-between">
								<dt class="text-sm text-muted-foreground">Shares Outstanding</dt>
								<dd class="font-mono font-medium tabular-nums">
									{fundamentals.sharesOutstanding ? formatVolume(fundamentals.sharesOutstanding) : '-'}
								</dd>
							</div>
						</dl>
					</div>
				{/if}

				<!-- Sector Peers -->
				{#if sectorPeers.length > 0}
					<div class="rounded-xl border bg-card">
						<div class="flex items-center gap-2 border-b px-6 py-4">
							<UsersIcon class="h-4 w-4 text-muted-foreground" />
							<h2 class="font-semibold">Sector Peers</h2>
						</div>
						<div class="divide-y">
							{#each sectorPeers.slice(0, 5) as peer}
								{@const peerPositive = peer.price && peer.price.change > 0}
								{@const peerNegative = peer.price && peer.price.change < 0}
								<a
									href="/company/{peer.company?.symbol}"
									class="flex items-center justify-between px-6 py-3 hover:bg-accent/50"
								>
									<div>
										<div class="font-mono font-medium">{peer.company?.symbol}</div>
										<div class="text-xs text-muted-foreground line-clamp-1">{peer.company?.name}</div>
									</div>
									<div class="text-right">
										<div class="font-mono text-sm tabular-nums">{peer.price?.ltp ? formatPriceCompact(peer.price.ltp) : '-'}</div>
										<div class="text-xs font-mono tabular-nums {peerPositive ? 'text-positive' : peerNegative ? 'text-negative' : 'text-muted-foreground'}">
											{peer.price?.percentChange !== undefined ? `${peer.price.percentChange >= 0 ? '+' : ''}${peer.price.percentChange.toFixed(2)}%` : '-'}
										</div>
									</div>
								</a>
							{/each}
						</div>
						<div class="border-t px-6 py-3">
							<a
								href="/companies?sector={company.sector}"
								class="text-sm text-muted-foreground hover:text-foreground"
							>
								View all {getSectorName(company.sector)} stocks →
							</a>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</PageContainer>
</section>
