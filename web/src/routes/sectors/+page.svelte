<script lang="ts">
	import { PageContainer } from '$lib/components/layout';
	import { SectorCard } from '$lib/components/market';
	import { formatVolume } from '$lib/utils/format';
	import { getSectorName } from '$lib/utils/sector';
	import { PieChartIcon } from '@lucide/svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const totalTurnover = $derived(
		data.sectors.reduce((sum, s) => sum + Number(s.turnover), 0)
	);

	const totalStocks = $derived(
		data.sectors.reduce((sum, s) => sum + s.stockCount, 0)
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
			<h1 class="text-2xl font-bold md:text-3xl">Sectors</h1>
			<p class="mt-1 text-muted-foreground">
				Explore {data.sectors.length} market sectors with {totalStocks} total stocks
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

		<!-- Sectors Grid -->
		{#if data.sectors.length > 0}
			<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
				{#each data.sectors as sector}
					<SectorCard
						sector={sector.sector}
						stockCount={sector.stockCount}
						turnover={sector.turnover}
					/>
				{/each}
			</div>
		{:else}
			<div class="rounded-lg border bg-muted/30 py-16 text-center">
				<PieChartIcon class="mx-auto h-12 w-12 text-muted-foreground/50" />
				<h3 class="mt-4 font-semibold">No sector data available</h3>
				<p class="mt-2 text-sm text-muted-foreground">
					Unable to load sector information at this time
				</p>
			</div>
		{/if}
	</PageContainer>
</section>
