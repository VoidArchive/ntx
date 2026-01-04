<script lang="ts">
	import { goto } from '$app/navigation';
	import { PageContainer } from '$lib/components/layout';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { getSectorName, getAllSectors, getSectorColor } from '$lib/utils/sector';
	import { formatPriceCompact, formatChange, formatVolume } from '$lib/utils/format';
	import { Sector } from '@ntx/api/ntx/v1/common_pb';
	import { SearchIcon, Building2Icon, XIcon } from '@lucide/svelte';
	import { debounce } from '$lib/utils/debounce';
	import type { PageData } from './$types';
	import { SvelteURLSearchParams } from 'svelte/reactivity';

	let { data }: { data: PageData } = $props();

	let searchInput = $derived(data.query);
	let sectorValue = $derived(data.sector !== Sector.UNSPECIFIED ? String(data.sector) : '');

	const sectors = getAllSectors();

	function updateFilters(newQuery?: string, newSector?: string) {
		const params = new SvelteURLSearchParams()
		const q = newQuery ?? searchInput;
		const s = newSector ?? sectorValue;

		if (q) params.set('q', q);
		if (s) params.set('sector', s);

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

	function onSectorChange(value: string | undefined) {
		sectorValue = value ?? '';
		updateFilters(undefined, value);
	}

	function clearFilters() {
		searchInput = '';
		sectorValue = '';
		goto('/companies', { invalidateAll: true });
	}

	const hasFilters = $derived(data.query || data.sector !== Sector.UNSPECIFIED);
</script>

<svelte:head>
	<title>Companies | NTX</title>
	<meta name="description" content="Browse all NEPSE listed companies with live prices. Filter by sector or search by symbol and name." />
</svelte:head>

<section class="py-8">
	<PageContainer>
		<!-- Header -->
		<div class="mb-8">
			<h1 class="text-2xl font-bold md:text-3xl">Companies</h1>
			<p class="mt-1 text-muted-foreground">
				{data.total} NEPSE listed companies
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
			<Select.Root type="single" value={sectorValue} onValueChange={onSectorChange}>
				<Select.Trigger class="w-full sm:w-48">
					{sectorValue ? getSectorName(parseInt(sectorValue, 10) as Sector) : 'All Sectors'}
				</Select.Trigger>
				<Select.Content>
					<Select.Item value="">All Sectors</Select.Item>
					{#each sectors as sector  }
						<Select.Item value={String(sector)}>{getSectorName(sector)}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>

		<!-- Active filters -->
		{#if hasFilters}
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
						<button onclick={() => onSectorChange('')} class="ml-1">
							<XIcon class="h-3 w-3" />
						</button>
					</Badge>
				{/if}
				<button onclick={clearFilters} class="text-xs text-muted-foreground hover:text-foreground">
					Clear all
				</button>
			</div>
		{/if}

		<!-- Companies Table -->
		{#if data.results.length > 0}
			<div class="rounded-lg border">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Company</Table.Head>
							<Table.Head>Sector</Table.Head>
							<Table.Head class="text-right">Price</Table.Head>
							<Table.Head class="text-right">Change</Table.Head>
							<Table.Head class="hidden text-right md:table-cell">Volume</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each data.results as result}
							{@const isPositive = result.price && result.price.change > 0}
							{@const isNegative = result.price && result.price.change < 0}
							<Table.Row 
								class="cursor-pointer hover:bg-accent/50" 
								onclick={() => goto(`/company/${result.company?.symbol}`)}
							>
								<Table.Cell>
									<div class="flex items-center gap-3">
										<div class="flex h-8 w-8 items-center justify-center rounded-md bg-primary/10 text-xs font-bold text-primary">
											{result.company?.symbol.slice(0, 2)}
										</div>
										<div>
											<div class="font-mono font-medium">{result.company?.symbol}</div>
											<div class="text-xs text-muted-foreground line-clamp-1 max-w-[200px]">{result.company?.name}</div>
										</div>
									</div>
								</Table.Cell>
								<Table.Cell>
									{#if result.company?.sector}
										<Badge variant="outline" class={getSectorColor(result.company.sector)}>
											{getSectorName(result.company.sector)}
										</Badge>
									{/if}
								</Table.Cell>
								<Table.Cell class="text-right font-mono tabular-nums">
									{result.price?.ltp ? formatPriceCompact(result.price.ltp) : '-'}
								</Table.Cell>
								<Table.Cell class="text-right font-mono tabular-nums {isPositive ? 'text-positive' : isNegative ? 'text-negative' : ''}">
									{result.price?.percentChange !== undefined ? formatChange(result.price.percentChange) : '-'}
								</Table.Cell>
								<Table.Cell class="hidden text-right font-mono tabular-nums md:table-cell">
									{result.price?.volume ? formatVolume(result.price.volume) : '-'}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
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
