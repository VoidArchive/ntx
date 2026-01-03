<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { PageContainer } from '$lib/components/layout';
	import { StockCard } from '$lib/components/stock';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import { Badge } from '$lib/components/ui/badge';
	import { getSectorName, getAllSectors } from '$lib/utils/sector';
	import { Sector } from '@ntx/api/ntx/v1/common_pb';
	import { SearchIcon, Building2Icon, XIcon } from '@lucide/svelte';
	import { debounce } from '$lib/utils/debounce';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let searchInput = $state(data.query);

	const sectors = getAllSectors();

	const selectedSector = $derived(
		data.sector !== Sector.UNSPECIFIED
			? { value: String(data.sector), label: getSectorName(data.sector) }
			: undefined
	);

	function updateFilters(newQuery?: string, newSector?: Sector) {
		const params = new URLSearchParams();
		const q = newQuery ?? searchInput;
		const s = newSector ?? data.sector;

		if (q) params.set('q', q);
		if (s !== Sector.UNSPECIFIED) params.set('sector', String(s));

		const search = params.toString();
		goto(search ? `?${search}` : '/companies', { invalidateAll: true });
	}

	const debouncedSearch = debounce((q: string) => {
		updateFilters(q);
	}, 300);

	function onSearchInput(e: Event) {
		const value = (e.target as HTMLInputElement).value;
		searchInput = value;
		debouncedSearch(value);
	}

	function onSectorChange(selected: { value: string; label: string } | undefined) {
		const sector = selected ? (parseInt(selected.value, 10) as Sector) : Sector.UNSPECIFIED;
		updateFilters(undefined, sector);
	}

	function clearFilters() {
		searchInput = '';
		goto('/companies', { invalidateAll: true });
	}
</script>

<svelte:head>
	<title>Companies | NTX</title>
	<meta name="description" content="Browse all NEPSE listed companies. Filter by sector or search by symbol and name." />
</svelte:head>

<section class="py-8">
	<PageContainer>
		<!-- Header -->
		<div class="mb-8">
			<h1 class="text-2xl font-bold md:text-3xl">Companies</h1>
			<p class="mt-1 text-muted-foreground">
				Browse all {data.companies.length} NEPSE listed companies
			</p>
		</div>

		<!-- Filters -->
		<div class="mb-6 flex flex-col gap-4 sm:flex-row">
			<div class="relative flex-1">
				<SearchIcon class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
				<Input
					type="text"
					placeholder="Search by symbol or name..."
					class="pl-9"
					value={searchInput}
					oninput={onSearchInput}
				/>
			</div>
			<Select.Root selected={selectedSector} onSelectedChange={onSectorChange}>
				<Select.Trigger class="w-full sm:w-48">
					<Select.Value placeholder="All Sectors" />
				</Select.Trigger>
				<Select.Content>
					<Select.Item value={String(Sector.UNSPECIFIED)}>All Sectors</Select.Item>
					{#each sectors as sector}
						<Select.Item value={String(sector)}>{getSectorName(sector)}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>

		<!-- Active filters -->
		{#if data.query || data.sector !== Sector.UNSPECIFIED}
			<div class="mb-6 flex flex-wrap items-center gap-2">
				<span class="text-sm text-muted-foreground">Filters:</span>
				{#if data.query}
					<Badge variant="secondary" class="flex items-center gap-1">
						Search: {data.query}
						<button onclick={() => { searchInput = ''; updateFilters(''); }} class="ml-1">
							<XIcon class="h-3 w-3" />
						</button>
					</Badge>
				{/if}
				{#if data.sector !== Sector.UNSPECIFIED}
					<Badge variant="secondary" class="flex items-center gap-1">
						{getSectorName(data.sector)}
						<button onclick={() => updateFilters(undefined, Sector.UNSPECIFIED)} class="ml-1">
							<XIcon class="h-3 w-3" />
						</button>
					</Badge>
				{/if}
				<button onclick={clearFilters} class="text-xs text-muted-foreground hover:text-foreground">
					Clear all
				</button>
			</div>
		{/if}

		<!-- Companies Grid -->
		{#if data.companies.length > 0}
			<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
				{#each data.companies as company}
					<StockCard
						symbol={company.symbol}
						name={company.name}
						price={0}
						change={0}
						percentChange={0}
						href="/company/{company.symbol}"
					/>
				{/each}
			</div>
		{:else}
			<div class="rounded-lg border bg-muted/30 py-16 text-center">
				<Building2Icon class="mx-auto h-12 w-12 text-muted-foreground/50" />
				<h3 class="mt-4 font-semibold">No companies found</h3>
				<p class="mt-2 text-sm text-muted-foreground">
					Try adjusting your search or filter criteria
				</p>
			</div>
		{/if}
	</PageContainer>
</section>
