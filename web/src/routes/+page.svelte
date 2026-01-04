<script lang="ts">
	import { goto } from '$app/navigation';
	import { PageContainer } from '$lib/components/layout';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { IndexCard, MarketStatusBadge, TopMoversCard } from '$lib/components/market';
	import { SearchIcon, SlidersHorizontalIcon, TrendingUpIcon, TrendingDownIcon, ArrowRightIcon } from '@lucide/svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	function openSearch() {
		window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true }));
	}
</script>

<svelte:head>
	<title>NTX - NEPSE Stock Screener</title>
	<meta name="description" content="Screen and filter NEPSE stocks by price, P/E ratio, and performance. Find top gainers, losers, and undervalued stocks." />
</svelte:head>

<!-- Hero Section -->
<section class="border-b bg-linear-to-b from-muted/50 to-background py-16 md:py-24">
	<PageContainer>
		<div class="mx-auto max-w-3xl text-center">
			{#if data.status}
				<div class="mb-6 flex justify-center">
					<MarketStatusBadge isOpen={data.status.isOpen} state={data.status.state} />
				</div>
			{/if}

			<h1 class="text-3xl font-bold tracking-tight md:text-4xl lg:text-5xl">
				Screen <span class="text-primary">NEPSE</span> stocks instantly
			</h1>
			<p class="mx-auto mt-4 max-w-xl text-muted-foreground">
				Filter by price, P/E ratio, 52-week range, and more. Find undervalued stocks, top movers, and sector leaders.
			</p>

			<!-- CTA Buttons -->
			<div class="mt-8 flex flex-wrap items-center justify-center gap-3">
				<Button size="lg" href="/screener">
					<SlidersHorizontalIcon class="mr-2 h-5 w-5" />
					Open Screener
				</Button>
				<Button variant="outline" size="lg" onclick={openSearch}>
					<SearchIcon class="mr-2 h-4 w-4" />
					Quick Search
					<kbd class="ml-2 hidden rounded border bg-muted px-1.5 py-0.5 font-mono text-xs sm:inline">
						Cmd+K
					</kbd>
				</Button>
			</div>

			<!-- Quick Filter Links -->
			<div class="mt-6 flex flex-wrap items-center justify-center gap-2">
				<span class="text-sm text-muted-foreground">Quick filters:</span>
				<Badge variant="outline" class="cursor-pointer hover:bg-accent" onclick={() => goto('/screener?near52wHigh=true')}>
					<TrendingUpIcon class="mr-1 h-3 w-3" />
					Near 52W High
				</Badge>
				<Badge variant="outline" class="cursor-pointer hover:bg-accent" onclick={() => goto('/screener?near52wLow=true')}>
					<TrendingDownIcon class="mr-1 h-3 w-3" />
					Near 52W Low
				</Badge>
				<Badge variant="outline" class="cursor-pointer hover:bg-accent" onclick={() => goto('/screener?maxPe=15')}>
					Low P/E (&lt;15)
				</Badge>
				<Badge variant="outline" class="cursor-pointer hover:bg-accent" onclick={() => goto('/screener?sort=volume&order=desc')}>
					High Volume
				</Badge>
			</div>
		</div>
	</PageContainer>
</section>

<!-- Market Indices -->
{#if data.indices && data.indices.length > 0}
	<section class="py-10">
		<PageContainer>
			<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
				{#each data.indices.slice(0, 4) as index (index.name)}
					<IndexCard
						name={index.name}
						value={index.value}
						change={index.change}
						percentChange={index.percentChange}
					/>
				{/each}
			</div>
		</PageContainer>
	</section>
{/if}

<!-- Top Movers -->
<section class="border-t bg-muted/30 py-10">
	<PageContainer>
		<div class="mb-6 flex items-center justify-between">
			<h2 class="text-lg font-semibold">Today's Movers</h2>
			<a href="/screener?sort=change&order=desc" class="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
				View all
				<ArrowRightIcon class="h-4 w-4" />
			</a>
		</div>
		<div class="grid gap-6 lg:grid-cols-2">
			<TopMoversCard
				title="Top Gainers"
				stocks={data.gainers ?? []}
				type="gainers"
				href="/screener?sort=change&order=desc"
			/>
			<TopMoversCard
				title="Top Losers"
				stocks={data.losers ?? []}
				type="losers"
				href="/screener?sort=change&order=asc"
			/>
		</div>
	</PageContainer>
</section>

<!-- Features -->
<section class="py-16">
	<PageContainer>
		<div class="mx-auto max-w-3xl text-center">
			<h2 class="text-2xl font-bold">Built for NEPSE Investors</h2>
			<p class="mt-2 text-muted-foreground">
				Everything you need to screen and analyze NEPSE stocks in one place
			</p>
		</div>

		<div class="mt-10 grid gap-6 md:grid-cols-3">
			<div class="rounded-xl border bg-card p-6">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
					<SlidersHorizontalIcon class="h-5 w-5 text-primary" />
				</div>
				<h3 class="mt-4 font-semibold">Advanced Screening</h3>
				<p class="mt-2 text-sm text-muted-foreground">
					Filter stocks by price range, P/E ratio, 52-week position, sector, and more. Find exactly what you're looking for.
				</p>
			</div>

			<div class="rounded-xl border bg-card p-6">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
					<TrendingUpIcon class="h-5 w-5 text-primary" />
				</div>
				<h3 class="mt-4 font-semibold">Real-time Data</h3>
				<p class="mt-2 text-sm text-muted-foreground">
					Live prices, volume, and market data synced throughout the trading day. Never miss a move.
				</p>
			</div>

			<div class="rounded-xl border bg-card p-6">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
					<SearchIcon class="h-5 w-5 text-primary" />
				</div>
				<h3 class="mt-4 font-semibold">Instant Search</h3>
				<p class="mt-2 text-sm text-muted-foreground">
					Press Cmd+K to instantly search any NEPSE company. Jump to detailed company pages with one click.
				</p>
			</div>
		</div>
	</PageContainer>
</section>
