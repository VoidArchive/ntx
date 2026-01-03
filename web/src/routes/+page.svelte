<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { PageContainer } from '$lib/components/layout';
	import { IndexCard, MarketStatusBadge, TopMoversCard, SectorCard } from '$lib/components/market';
	import { SearchIcon, GithubIcon, ArrowRightIcon } from '@lucide/svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	function openSearch() {
		// Dispatch keyboard event to trigger search dialog in layout
		window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true }));
	}
</script>

<!-- Hero Section -->
<section class="relative overflow-hidden border-b bg-gradient-to-b from-muted/50 to-background py-20 md:py-32">
	<PageContainer>
		<div class="mx-auto max-w-3xl text-center">
			<!-- Market Status -->
			{#if data.status}
				<div class="mb-6 flex justify-center">
					<MarketStatusBadge isOpen={data.status.isOpen} state={data.status.state} />
				</div>
			{/if}

			<h1 class="text-4xl font-bold tracking-tight md:text-5xl lg:text-6xl">
				Find any <span class="text-primary">NEPSE</span> stock instantly
			</h1>
			<p class="mx-auto mt-4 max-w-xl text-lg text-muted-foreground">
				Open-source stock data aggregator with screening capabilities. Access company fundamentals,
				price history, and market insights.
			</p>

			<!-- Search Bar -->
			<div class="mx-auto mt-8 max-w-xl">
				<button
					onclick={openSearch}
					class="group flex w-full items-center gap-3 rounded-xl border bg-background px-4 py-4 text-left shadow-sm transition-all hover:border-primary/50 hover:shadow-md"
				>
					<SearchIcon class="h-5 w-5 text-muted-foreground" />
					<span class="flex-1 text-muted-foreground">Search companies by symbol or name...</span>
					<kbd
						class="hidden rounded border bg-muted px-2 py-1 font-mono text-xs text-muted-foreground sm:inline-block"
					>
						⌘K
					</kbd>
				</button>
			</div>

			<!-- Quick links -->
			<div class="mt-6 flex flex-wrap items-center justify-center gap-3">
				<Button variant="outline" size="sm" href="/companies">
					Browse Companies
					<ArrowRightIcon class="ml-1 h-4 w-4" />
				</Button>
				<Button variant="outline" size="sm" href="/screener">
					Stock Screener
					<ArrowRightIcon class="ml-1 h-4 w-4" />
				</Button>
				<Button variant="ghost" size="sm" href="https://github.com/your-repo/ntx" target="_blank">
					<GithubIcon class="mr-1 h-4 w-4" />
					Star on GitHub
				</Button>
			</div>
		</div>
	</PageContainer>
</section>

<!-- Market Indices -->
{#if data.indices && data.indices.length > 0}
	<section class="py-12">
		<PageContainer>
			<h2 class="mb-6 text-lg font-semibold">Market Indices</h2>
			<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
				{#each data.indices.slice(0, 4) as index}
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
<section class="border-t bg-muted/30 py-12">
	<PageContainer>
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

<!-- Sectors -->
{#if data.sectors && data.sectors.length > 0}
	<section class="py-12">
		<PageContainer>
			<div class="mb-6 flex items-center justify-between">
				<h2 class="text-lg font-semibold">Sectors</h2>
				<a
					href="/sectors"
					class="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
				>
					View all
					<ArrowRightIcon class="h-4 w-4" />
				</a>
			</div>
			<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
				{#each data.sectors.slice(0, 8) as sector}
					<SectorCard
						sector={sector.sector}
						stockCount={sector.stockCount}
						turnover={sector.turnover}
					/>
				{/each}
			</div>
		</PageContainer>
	</section>
{/if}

<!-- Open Source CTA -->
<section class="border-t bg-muted/30 py-16">
	<PageContainer>
		<div class="mx-auto max-w-2xl text-center">
			<div
				class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10"
			>
				<GithubIcon class="h-6 w-6 text-primary" />
			</div>
			<h2 class="text-2xl font-bold">Open Source</h2>
			<p class="mt-2 text-muted-foreground">
				NTX is free and open source. Contribute, report issues, or fork for your own projects.
			</p>
			<div class="mt-6 flex flex-wrap items-center justify-center gap-3">
				<Button href="https://github.com/your-repo/ntx" target="_blank">
					<GithubIcon class="mr-2 h-4 w-4" />
					View on GitHub
				</Button>
				<Button variant="outline" href="https://github.com/your-repo/ntx/issues" target="_blank">
					Report an Issue
				</Button>
			</div>
		</div>
	</PageContainer>
</section>
