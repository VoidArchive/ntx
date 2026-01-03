<script lang="ts">
	import { PageContainer } from '$lib/components/layout';
	import { PriceDisplay, FundamentalsGrid, ReportTable } from '$lib/components/stock';
	import { Badge } from '$lib/components/ui/badge';
	import * as Tabs from '$lib/components/ui/tabs';
	import { getSectorName, getSectorColor } from '$lib/utils/sector';
	import { formatVolume, formatPriceCompact, formatDate } from '$lib/utils/format';
	import {
		TrendingUpIcon,
		TrendingDownIcon,
		CalendarIcon,
		BarChart3Icon,
		FileTextIcon,
		InfoIcon
	} from '@lucide/svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const { company, price: stockPrice, fundamentals, reports } = data;
</script>

<svelte:head>
	<title>{company.symbol} - {company.name} | NTX</title>
	<meta
		name="description"
		content="View {company.name} ({company.symbol}) stock price, fundamentals, and financial data on NTX."
	/>
</svelte:head>

<!-- Company Header -->
<section class="border-b bg-muted/30 py-8">
	<PageContainer>
		<div class="flex flex-col gap-6 md:flex-row md:items-start md:justify-between">
			<!-- Company Info -->
			<div>
				<div class="flex items-center gap-3">
					<div
						class="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-lg font-bold text-primary"
					>
						{company.symbol.slice(0, 2)}
					</div>
					<div>
						<h1 class="flex items-center gap-2 text-2xl font-bold md:text-3xl">
							{company.symbol}
							<Badge variant="secondary" class={getSectorColor(company.sector)}>
								{getSectorName(company.sector)}
							</Badge>
						</h1>
						<p class="text-muted-foreground">{company.name}</p>
					</div>
				</div>
			</div>

			<!-- Price Info -->
			{#if stockPrice}
				<div class="md:text-right">
					<PriceDisplay
						price={stockPrice.ltp}
						change={stockPrice.change}
						percentChange={stockPrice.percentChange}
						size="lg"
					/>
					{#if stockPrice.timestamp}
						<p class="mt-1 text-xs text-muted-foreground">
							Last updated: {formatDate(new Date(Number(stockPrice.timestamp.seconds) * 1000))}
						</p>
					{/if}
				</div>
			{/if}
		</div>

		<!-- Quick Stats -->
		{#if stockPrice}
			<div class="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-4 lg:grid-cols-6">
				<div>
					<div class="text-xs font-medium uppercase text-muted-foreground">Open</div>
					<div class="font-mono font-medium tabular-nums">{formatPriceCompact(stockPrice.open)}</div>
				</div>
				<div>
					<div class="text-xs font-medium uppercase text-muted-foreground">High</div>
					<div class="font-mono font-medium tabular-nums text-positive">
						{formatPriceCompact(stockPrice.high)}
					</div>
				</div>
				<div>
					<div class="text-xs font-medium uppercase text-muted-foreground">Low</div>
					<div class="font-mono font-medium tabular-nums text-negative">
						{formatPriceCompact(stockPrice.low)}
					</div>
				</div>
				<div>
					<div class="text-xs font-medium uppercase text-muted-foreground">Prev. Close</div>
					<div class="font-mono font-medium tabular-nums">
						{formatPriceCompact(stockPrice.previousClose)}
					</div>
				</div>
				<div>
					<div class="text-xs font-medium uppercase text-muted-foreground">Volume</div>
					<div class="font-mono font-medium tabular-nums">{formatVolume(stockPrice.volume)}</div>
				</div>
				<div>
					<div class="text-xs font-medium uppercase text-muted-foreground">Turnover</div>
					<div class="font-mono font-medium tabular-nums">{formatVolume(stockPrice.turnover)}</div>
				</div>
			</div>
		{/if}
	</PageContainer>
</section>

<!-- Tabs Content -->
<section class="py-8">
	<PageContainer>
		<Tabs.Root value="overview" class="w-full">
			<Tabs.List class="mb-6 w-full justify-start">
				<Tabs.Trigger value="overview" class="flex items-center gap-2">
					<InfoIcon class="h-4 w-4" />
					<span class="hidden sm:inline">Overview</span>
				</Tabs.Trigger>
				<Tabs.Trigger value="fundamentals" class="flex items-center gap-2">
					<BarChart3Icon class="h-4 w-4" />
					<span class="hidden sm:inline">Fundamentals</span>
				</Tabs.Trigger>
				<Tabs.Trigger value="financials" class="flex items-center gap-2">
					<FileTextIcon class="h-4 w-4" />
					<span class="hidden sm:inline">Financials</span>
				</Tabs.Trigger>
				<Tabs.Trigger value="history" class="flex items-center gap-2">
					<CalendarIcon class="h-4 w-4" />
					<span class="hidden sm:inline">History</span>
				</Tabs.Trigger>
			</Tabs.List>

			<!-- Overview Tab -->
			<Tabs.Content value="overview">
				<div class="grid gap-6 lg:grid-cols-3">
					<!-- Main content -->
					<div class="lg:col-span-2">
						{#if company.description}
							<div class="rounded-lg border bg-card p-6">
								<h3 class="font-semibold">About {company.name}</h3>
								<p class="mt-2 text-sm text-muted-foreground leading-relaxed">
									{company.description}
								</p>
							</div>
						{/if}

						<!-- 52-Week Range -->
						{#if stockPrice}
							<div class="mt-6 rounded-lg border bg-card p-6">
								<h3 class="font-semibold">52-Week Range</h3>
								<div class="mt-4">
									<div class="flex justify-between text-sm">
										<span class="text-negative">Low: {formatPriceCompact(stockPrice.week52Low)}</span>
										<span class="text-positive">High: {formatPriceCompact(stockPrice.week52High)}</span>
									</div>
									<div class="relative mt-2 h-2 rounded-full bg-muted">
										{@const range = stockPrice.week52High - stockPrice.week52Low}
										{@const position = range > 0 ? ((stockPrice.ltp - stockPrice.week52Low) / range) * 100 : 50}
										<div
											class="absolute top-0 h-2 rounded-full bg-primary"
											style="width: {position}%"
										></div>
										<div
											class="absolute top-1/2 h-4 w-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-primary bg-background"
											style="left: {position}%"
										></div>
									</div>
									<div class="mt-2 text-center text-sm text-muted-foreground">
										Current: Rs. {formatPriceCompact(stockPrice.ltp)}
									</div>
								</div>
							</div>
						{/if}
					</div>

					<!-- Sidebar with key stats -->
					<div>
						{#if fundamentals}
							<div class="rounded-lg border bg-card p-6">
								<h3 class="font-semibold">Key Statistics</h3>
								<dl class="mt-4 space-y-3">
									<div class="flex justify-between">
										<dt class="text-sm text-muted-foreground">P/E Ratio</dt>
										<dd class="font-mono font-medium tabular-nums">
											{fundamentals.pe ? fundamentals.pe.toFixed(2) : '-'}
										</dd>
									</div>
									<div class="flex justify-between">
										<dt class="text-sm text-muted-foreground">P/B Ratio</dt>
										<dd class="font-mono font-medium tabular-nums">
											{fundamentals.pb ? fundamentals.pb.toFixed(2) : '-'}
										</dd>
									</div>
									<div class="flex justify-between">
										<dt class="text-sm text-muted-foreground">EPS</dt>
										<dd class="font-mono font-medium tabular-nums">
											{fundamentals.eps ? `Rs. ${fundamentals.eps.toFixed(2)}` : '-'}
										</dd>
									</div>
									<div class="flex justify-between">
										<dt class="text-sm text-muted-foreground">ROE</dt>
										<dd class="font-mono font-medium tabular-nums">
											{fundamentals.roe ? `${fundamentals.roe.toFixed(2)}%` : '-'}
										</dd>
									</div>
									<div class="flex justify-between">
										<dt class="text-sm text-muted-foreground">Dividend Yield</dt>
										<dd class="font-mono font-medium tabular-nums">
											{fundamentals.dividendYield ? `${fundamentals.dividendYield.toFixed(2)}%` : '-'}
										</dd>
									</div>
								</dl>
							</div>
						{/if}
					</div>
				</div>
			</Tabs.Content>

			<!-- Fundamentals Tab -->
			<Tabs.Content value="fundamentals">
				{#if fundamentals}
					<FundamentalsGrid {fundamentals} />
				{:else}
					<div class="rounded-lg border bg-muted/30 py-12 text-center">
						<p class="text-muted-foreground">No fundamentals data available for {company.symbol}</p>
					</div>
				{/if}
			</Tabs.Content>

			<!-- Financials Tab -->
			<Tabs.Content value="financials">
				<ReportTable {reports} />
			</Tabs.Content>

			<!-- History Tab -->
			<Tabs.Content value="history">
				<div class="rounded-lg border bg-muted/30 py-12 text-center">
					<BarChart3Icon class="mx-auto h-12 w-12 text-muted-foreground/50" />
					<h3 class="mt-4 font-semibold">Price History Chart</h3>
					<p class="mt-2 text-sm text-muted-foreground">
						Interactive price chart coming soon. Use the Overview tab to see 52-week range.
					</p>
				</div>
			</Tabs.Content>
		</Tabs.Root>
	</PageContainer>
</section>
