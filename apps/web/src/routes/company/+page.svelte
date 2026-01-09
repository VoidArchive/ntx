<script lang="ts">
	import CompanyCard from '$lib/components/CompanyCard.svelte';
	import { Sector } from '$lib/gen/ntx/v1/common_pb';

	let { data } = $props();

	let selectedSector = $state<number | null>(null);
	let searchQuery = $state('');

	const sectors = [
		{ value: Sector.COMMERCIAL_BANK, label: 'Banks' },
		{ value: Sector.DEVELOPMENT_BANK, label: 'Dev Banks' },
		{ value: Sector.FINANCE, label: 'Finance' },
		{ value: Sector.MICROFINANCE, label: 'Microfinance' },
		{ value: Sector.LIFE_INSURANCE, label: 'Life Insurance' },
		{ value: Sector.NON_LIFE_INSURANCE, label: 'Non-Life Insurance' },
		{ value: Sector.HYDROPOWER, label: 'Hydropower' },
		{ value: Sector.MANUFACTURING, label: 'Manufacturing' },
		{ value: Sector.HOTEL, label: 'Hotels' },
		{ value: Sector.TRADING, label: 'Trading' },
		{ value: Sector.INVESTMENT, label: 'Investment' },
		{ value: Sector.MUTUAL_FUND, label: 'Mutual Fund' },
		{ value: Sector.OTHERS, label: 'Others' }
	];

	let filteredCompanies = $derived.by(() => {
		let companies = data.companies ?? [];

		if (selectedSector !== null) {
			companies = companies.filter((c) => c.sector === selectedSector);
		}

		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase();
			companies = companies.filter(
				(c) =>
					c.symbol.toLowerCase().includes(query) || c.name.toLowerCase().includes(query)
			);
		}

		return companies;
	});

	function toggleSector(sector: number) {
		selectedSector = selectedSector === sector ? null : sector;
	}

	function clearFilters() {
		selectedSector = null;
		searchQuery = '';
	}

	let hasFilters = $derived(selectedSector !== null || searchQuery.trim().length > 0);
</script>

<div class="min-h-screen">
	<!-- Header -->
	<header class="border-b border-border">
		<div class="mx-auto max-w-3xl px-4 py-8">
			<h1 class="text-3xl tracking-tight">Companies</h1>
			<p class="mt-1 text-muted-foreground">
				{data.companies?.length ?? 0} listed on NEPSE
			</p>
		</div>
	</header>

	<!-- Filters -->
	<div class="sticky top-0 z-40 border-b border-border bg-background">
		<div class="mx-auto max-w-3xl px-4 py-3">
			<!-- Search -->
			<div class="relative">
				<input
					type="search"
					placeholder="Search companies..."
					bind:value={searchQuery}
					class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm transition-colors placeholder:text-muted-foreground focus:border-foreground focus:outline-none"
				/>
			</div>

			<!-- Sector filters -->
			<div class="mt-3 flex flex-wrap gap-1.5">
				{#each sectors as sector (sector.value)}
					<button
						onclick={() => toggleSector(sector.value)}
						class="rounded-full px-2.5 py-1 text-xs transition-colors
						{selectedSector === sector.value
							? 'bg-foreground text-background'
							: 'bg-muted text-muted-foreground hover:bg-accent hover:text-foreground'}"
					>
						{sector.label}
					</button>
				{/each}
			</div>

			<!-- Results & Clear -->
			{#if hasFilters}
				<div class="mt-3 flex items-center justify-between text-sm">
					<span class="text-muted-foreground">
						{filteredCompanies.length} result{filteredCompanies.length === 1 ? '' : 's'}
					</span>
					<button onclick={clearFilters} class="text-muted-foreground hover:text-foreground">
						Clear
					</button>
				</div>
			{/if}
		</div>
	</div>

	<!-- Company List -->
	<main class="mx-auto max-w-3xl px-4">
		{#if filteredCompanies.length === 0}
			<div class="py-16 text-center">
				<p class="text-muted-foreground">No companies found</p>
				<button onclick={clearFilters} class="mt-2 text-sm hover:underline">
					Clear filters
				</button>
			</div>
		{:else}
			<div class="divide-y divide-border">
				{#each filteredCompanies as company (company.id)}
					<CompanyCard {company} />
				{/each}
			</div>
		{/if}
	</main>
</div>
