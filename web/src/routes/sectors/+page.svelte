<script lang="ts">
	import { goto } from '$app/navigation';
	import { PageContainer } from '$lib/components/layout';
	import { Badge } from '$lib/components/ui/badge';
	import { formatVolume } from '$lib/utils/format';
	import { getSectorName, getSectorColor } from '$lib/utils/sector';
	import { TrendingUpIcon, TrendingDownIcon, MinusIcon, Building2Icon, ArrowRightIcon } from '@lucide/svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const totalTurnover = $derived(
		data.sectors.reduce((sum, s) => sum + Number(s.turnover), 0)
	);

	const totalStocks = $derived(
		data.sectors.reduce((sum, s) => sum + s.stockCount, 0)
	);

	// Sort sectors by turnover for display
	const sortedSectors = $derived(
		[...data.sectors].sort((a, b) => Number(b.turnover) - Number(a.turnover))
	);
</script>

<svelte:head>
	<title>Sectors | NTX</title>
	<meta name="description" content="Explore NEPSE market sectors. View sector-wise turnover, stock counts, and performance." />
</svelte:head>

<section class="py-8">
	<PageContainer>
		<!-- Header -->
		<div class="mb-8">
			<h1 class="text-2xl font-bold md:text-3xl">Market Sectors</h1>
			<p class="mt-1 text-muted-foreground">
				{data.sectors.length} sectors with {totalStocks} total stocks
			</p>
		</div>

		<!-- Summary Stats -->
		<div class="mb-8 grid gap-4 sm:grid-cols-3">
			<div class="rounded-xl border bg-card p-6">
				<div class="text-sm font-medium text-muted-foreground">Total Sectors</div>
				<div class="mt-1 text-3xl font-bold">{data.sectors.length}</div>
			</div>
			<div class="rounded-xl border bg-card p-6">
				<div class="text-sm font-medium text-muted-foreground">Listed Companies</div>
				<div class="mt-1 text-3xl font-bold">{totalStocks}</div>
			</div>
			<div class="rounded-xl border bg-card p-6">
				<div class="text-sm font-medium text-muted-foreground">Total Turnover</div>
				<div class="mt-1 text-3xl font-bold">{formatVolume(totalTurnover)}</div>
			</div>
		</div>

		<!-- Sectors List -->
		{#if sortedSectors.length > 0}
			<div class="space-y-3">
				{#each sortedSectors as sector}
					{@const isPositive = sector.change > 0}
					{@const isNegative = sector.change < 0}
					{@const turnoverPercent = totalTurnover > 0 ? (Number(sector.turnover) / totalTurnover) * 100 : 0}
					<a
						href="/companies?sector={sector.sector}"
						class="group flex flex-col rounded-xl border bg-card p-4 transition-colors hover:bg-accent/50 sm:flex-row sm:items-center sm:justify-between"
					>
						<div class="flex items-center gap-4">
							<div class="rounded-lg p-2.5 {getSectorColor(sector.sector)}">
								<Building2Icon class="h-5 w-5" />
							</div>
							<div>
								<div class="font-semibold">{getSectorName(sector.sector)}</div>
								<div class="text-sm text-muted-foreground">
									{sector.stockCount} companies
								</div>
							</div>
						</div>

						<div class="mt-4 flex items-center justify-between gap-8 sm:mt-0">
							<!-- Performance -->
							<div class="text-right">
								<div class="text-xs text-muted-foreground">Performance</div>
								<div class="flex items-center justify-end gap-1 font-mono font-medium tabular-nums {isPositive ? 'text-positive' : isNegative ? 'text-negative' : 'text-muted-foreground'}">
									{#if isPositive}
										<TrendingUpIcon class="h-4 w-4" />
									{:else if isNegative}
										<TrendingDownIcon class="h-4 w-4" />
									{:else}
										<MinusIcon class="h-4 w-4" />
									{/if}
									{sector.percentChange >= 0 ? '+' : ''}{sector.percentChange.toFixed(2)}%
								</div>
							</div>

							<!-- Turnover -->
							<div class="text-right min-w-[100px]">
								<div class="text-xs text-muted-foreground">Turnover</div>
								<div class="font-mono font-medium tabular-nums">{formatVolume(sector.turnover)}</div>
								<div class="mt-1 h-1.5 w-full rounded-full bg-muted overflow-hidden">
									<div 
										class="h-full rounded-full bg-primary transition-all"
										style="width: {Math.min(turnoverPercent * 2, 100)}%"
									></div>
								</div>
							</div>

							<!-- Arrow -->
							<ArrowRightIcon class="h-5 w-5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100" />
						</div>
					</a>
				{/each}
			</div>
		{:else}
			<div class="rounded-lg border bg-muted/30 py-16 text-center">
				<Building2Icon class="mx-auto h-12 w-12 text-muted-foreground/50" />
				<h3 class="mt-4 font-semibold">No sector data available</h3>
				<p class="mt-2 text-sm text-muted-foreground">
					Unable to load sector information at this time
				</p>
			</div>
		{/if}
	</PageContainer>
</section>
