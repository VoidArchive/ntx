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

	import Search from '@lucide/svelte/icons/search';

	function toggleSector(sector: number) {
		selectedSector = selectedSector === sector ? null : sector;
	}

	function clearFilters() {
		selectedSector = null;
		searchQuery = '';
	}

	let hasFilters = $derived(selectedSector !== null || searchQuery.trim().length > 0);
	let searchInput: HTMLInputElement;

	function handleKeydown(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
			e.preventDefault();
			searchInput?.focus();
		}
	}

	$effect(() => {
		window.addEventListener('keydown', handleKeydown);
		return () => window.removeEventListener('keydown', handleKeydown);
	});
</script>

<div class="min-h-screen pb-20">
	<!-- Hero Section -->
	<div class="relative overflow-hidden border-b border-border bg-background/50 py-24 md:py-32">
		<div class="absolute inset-0 -z-10 opacity-30">
			<div class="absolute inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-primary/20 via-background to-background"></div>
		</div>

		<div class="mx-auto max-w-5xl px-4 text-center">
			<h1 class="font-serif text-5xl font-medium tracking-tight sm:text-7xl">
				Market <span class="italic text-muted-foreground">Stories</span>
			</h1>
			<p class="mx-auto mt-6 max-w-2xl text-lg text-muted-foreground">
				Discover the narrative behind every stock. Deep fundamental analysis, technical insights, and AI-driven research for {data.companies?.length ?? 0} NEPSE companies.
			</p>

			<!-- Search Bar -->
			<div class="mx-auto mt-10 max-w-xl">
				<div class="relative group">
					<div class="absolute -inset-0.5 rounded-full bg-gradient-to-r from-primary/20 to-secondary/20 opacity-50 blur transition duration-1000 group-hover:opacity-100 group-hover:duration-200"></div>
					<div class="relative flex items-center rounded-full bg-background ring-1 ring-border shadow-sm">
						<input
							bind:this={searchInput}
							type="search"
							placeholder="Search by symbol or name (e.g. NABIL)..."
							bind:value={searchQuery}
							class="flex-1 border-none bg-transparent px-6 py-4 text-base placeholder:text-muted-foreground focus:ring-0 focus:outline-none"
						/>
						<button 
							onclick={() => searchInput?.focus()}
							class="pr-4 text-muted-foreground hover:text-foreground transition-colors"
						>
							<Search class="size-5" />
						</button>
					</div>
				</div>
			</div>
			
			<!-- Scrollable Sector Filter Bar -->
			<div class="mt-8 flex flex-wrap justify-center gap-2">
				{#each sectors as sector (sector.value)}
					<button
						onclick={() => toggleSector(sector.value)}
						class="rounded-full border px-4 py-1.5 text-xs font-medium transition-all
						{selectedSector === sector.value
							? 'border-foreground bg-foreground text-background shadow-md'
							: 'border-border bg-background/50 text-muted-foreground hover:border-foreground/50 hover:text-foreground'}"
					>
						{sector.label}
					</button>
				{/each}
			</div>
		</div>
	</div>

	<!-- Results Section -->
	<main class="mx-auto max-w-7xl px-4 py-12">
		{#if hasFilters}
			<div class="mb-8 flex items-center justify-between">
				<h2 class="text-lg font-medium">
					{filteredCompanies.length} Result{filteredCompanies.length === 1 ? '' : 's'}
				</h2>
				<button onclick={clearFilters} class="text-sm text-muted-foreground hover:text-foreground hover:underline">
					Clear Filters
				</button>
			</div>
		{/if}

		{#if filteredCompanies.length === 0}
			<div class="rounded-2xl border border-dashed border-border p-16 text-center">
				<p class="text-lg text-muted-foreground">No companies match your search.</p>
				<button onclick={clearFilters} class="mt-4 text-sm font-medium hover:underline">
					Clear all filters
				</button>
			</div>
		{:else}
			<div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
				{#each filteredCompanies as company (company.id)}
					{@const price = data.prices?.find(p => p.companyId === company.id)}
					<CompanyCard {company} {price} />
				{/each}
			</div>
		{/if}
	</main>

</div>
