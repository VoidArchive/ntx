<script lang="ts">
	import { goto } from '$app/navigation';
	import { PageContainer } from '$lib/components/layout';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import * as Select from '$lib/components/ui/select';
	import * as Table from '$lib/components/ui/table';
	import { getSectorName, getAllSectors } from '$lib/utils/sector';
	import { formatPriceCompact, formatChange, formatVolume, formatMarketCap } from '$lib/utils/format';
	import { Sector } from '@ntx/api/ntx/v1/common_pb';
	import {
		SlidersHorizontalIcon,
		TrendingUpIcon,
		TrendingDownIcon,
		XIcon,
		ChevronLeftIcon,
		ChevronRightIcon
	} from '@lucide/svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const sectors = getAllSectors();

	// Local filter state
	let minPrice = $state(data.filters.minPrice?.toString() ?? '');
	let maxPrice = $state(data.filters.maxPrice?.toString() ?? '');
	let minPe = $state(data.filters.minPe?.toString() ?? '');
	let maxPe = $state(data.filters.maxPe?.toString() ?? '');
	let near52wHigh = $state(data.filters.near52wHigh ?? false);
	let near52wLow = $state(data.filters.near52wLow ?? false);

	const selectedSector = $derived(
		data.filters.sector !== Sector.UNSPECIFIED
			? { value: String(data.filters.sector), label: getSectorName(data.filters.sector as Sector) }
			: undefined
	);

	const selectedSort = $derived(
		data.filters.sortBy
			? { value: data.filters.sortBy, label: sortOptions.find((o) => o.value === data.filters.sortBy)?.label ?? '' }
			: undefined
	);

	const sortOptions = [
		{ value: 'change', label: 'Change %' },
		{ value: 'price', label: 'Price' },
		{ value: 'volume', label: 'Volume' },
		{ value: 'turnover', label: 'Turnover' },
		{ value: 'marketCap', label: 'Market Cap' },
		{ value: 'pe', label: 'P/E Ratio' },
		{ value: 'symbol', label: 'Symbol' }
	];

	function buildUrl(overrides: Record<string, string | number | boolean | undefined> = {}) {
		const params = new URLSearchParams();

		const values = {
			sector: data.filters.sector,
			minPrice,
			maxPrice,
			minPe,
			maxPe,
			near52wHigh,
			near52wLow,
			sort: data.filters.sortBy,
			order: data.filters.sortOrder,
			limit: data.filters.limit,
			offset: data.filters.offset,
			...overrides
		};

		if (values.sector && values.sector !== Sector.UNSPECIFIED) params.set('sector', String(values.sector));
		if (values.minPrice) params.set('minPrice', values.minPrice.toString());
		if (values.maxPrice) params.set('maxPrice', values.maxPrice.toString());
		if (values.minPe) params.set('minPe', values.minPe.toString());
		if (values.maxPe) params.set('maxPe', values.maxPe.toString());
		if (values.near52wHigh) params.set('near52wHigh', 'true');
		if (values.near52wLow) params.set('near52wLow', 'true');
		if (values.sort) params.set('sort', values.sort.toString());
		if (values.order) params.set('order', values.order.toString());
		if (values.offset && values.offset > 0) params.set('offset', values.offset.toString());

		const search = params.toString();
		return search ? `/screener?${search}` : '/screener';
	}

	function applyFilters() {
		goto(buildUrl({ offset: 0 }), { invalidateAll: true });
	}

	function clearFilters() {
		minPrice = '';
		maxPrice = '';
		minPe = '';
		maxPe = '';
		near52wHigh = false;
		near52wLow = false;
		goto('/screener', { invalidateAll: true });
	}

	function onSectorChange(selected: { value: string } | undefined) {
		const sector = selected ? parseInt(selected.value, 10) : Sector.UNSPECIFIED;
		goto(buildUrl({ sector, offset: 0 }), { invalidateAll: true });
	}

	function onSortChange(selected: { value: string } | undefined) {
		goto(buildUrl({ sort: selected?.value, offset: 0 }), { invalidateAll: true });
	}

	function toggleSortOrder() {
		const newOrder = data.filters.sortOrder === 'asc' ? 'desc' : 'asc';
		goto(buildUrl({ order: newOrder }), { invalidateAll: true });
	}

	const hasActiveFilters = $derived(
		minPrice || maxPrice || minPe || maxPe || near52wHigh || near52wLow || data.filters.sector !== Sector.UNSPECIFIED
	);

	const currentPage = $derived(Math.floor((data.filters.offset ?? 0) / (data.filters.limit ?? 50)) + 1);
	const totalPages = $derived(Math.ceil(data.total / (data.filters.limit ?? 50)));
</script>

<svelte:head>
	<title>Stock Screener | NTX</title>
	<meta name="description" content="Screen NEPSE stocks by fundamentals, price, and performance metrics." />
</svelte:head>

<section class="py-8">
	<PageContainer>
		<!-- Header -->
		<div class="mb-8">
			<h1 class="text-2xl font-bold md:text-3xl">Stock Screener</h1>
			<p class="mt-1 text-muted-foreground">
				Filter stocks by fundamentals and performance metrics
			</p>
		</div>

		<!-- Filters Panel -->
		<div class="mb-6 rounded-xl border bg-card p-4 md:p-6">
			<div class="flex items-center gap-2 mb-4">
				<SlidersHorizontalIcon class="h-5 w-5 text-muted-foreground" />
				<h2 class="font-semibold">Filters</h2>
				{#if hasActiveFilters}
					<button onclick={clearFilters} class="ml-auto text-xs text-muted-foreground hover:text-foreground">
						Clear all
					</button>
				{/if}
			</div>

			<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
				<!-- Sector -->
				<div>
					<label class="mb-1.5 block text-sm font-medium">Sector</label>
					<Select.Root selected={selectedSector} onSelectedChange={onSectorChange}>
						<Select.Trigger>
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

				<!-- Price Range -->
				<div>
					<label class="mb-1.5 block text-sm font-medium">Price Range</label>
					<div class="flex gap-2">
						<Input type="number" placeholder="Min" bind:value={minPrice} class="w-full" />
						<Input type="number" placeholder="Max" bind:value={maxPrice} class="w-full" />
					</div>
				</div>

				<!-- P/E Range -->
				<div>
					<label class="mb-1.5 block text-sm font-medium">P/E Ratio</label>
					<div class="flex gap-2">
						<Input type="number" placeholder="Min" bind:value={minPe} class="w-full" />
						<Input type="number" placeholder="Max" bind:value={maxPe} class="w-full" />
					</div>
				</div>

				<!-- Quick Filters -->
				<div>
					<label class="mb-1.5 block text-sm font-medium">Quick Filters</label>
					<div class="flex flex-wrap gap-2">
						<Badge
							variant={near52wHigh ? 'default' : 'outline'}
							class="cursor-pointer"
							onclick={() => (near52wHigh = !near52wHigh)}
						>
							<TrendingUpIcon class="mr-1 h-3 w-3" />
							Near 52W High
						</Badge>
						<Badge
							variant={near52wLow ? 'default' : 'outline'}
							class="cursor-pointer"
							onclick={() => (near52wLow = !near52wLow)}
						>
							<TrendingDownIcon class="mr-1 h-3 w-3" />
							Near 52W Low
						</Badge>
					</div>
				</div>
			</div>

			<div class="mt-4 flex justify-end">
				<Button onclick={applyFilters}>Apply Filters</Button>
			</div>
		</div>

		<!-- Sort Controls -->
		<div class="mb-4 flex flex-wrap items-center justify-between gap-4">
			<div class="text-sm text-muted-foreground">
				Found <span class="font-medium text-foreground">{data.total}</span> results
			</div>
			<div class="flex items-center gap-2">
				<Select.Root selected={selectedSort} onSelectedChange={onSortChange}>
					<Select.Trigger class="w-36">
						<Select.Value placeholder="Sort by" />
					</Select.Trigger>
					<Select.Content>
						{#each sortOptions as option}
							<Select.Item value={option.value}>{option.label}</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
				<Button variant="outline" size="icon" onclick={toggleSortOrder}>
					{#if data.filters.sortOrder === 'asc'}
						<TrendingUpIcon class="h-4 w-4" />
					{:else}
						<TrendingDownIcon class="h-4 w-4" />
					{/if}
				</Button>
			</div>
		</div>

		<!-- Results Table -->
		{#if data.results.length > 0}
			<div class="rounded-lg border">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Symbol</Table.Head>
							<Table.Head class="text-right">Price</Table.Head>
							<Table.Head class="text-right">Change</Table.Head>
							<Table.Head class="hidden text-right md:table-cell">Volume</Table.Head>
							<Table.Head class="hidden text-right lg:table-cell">Market Cap</Table.Head>
							<Table.Head class="hidden text-right lg:table-cell">P/E</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each data.results as result}
							{@const isPositive = result.price && result.price.change > 0}
							{@const isNegative = result.price && result.price.change < 0}
							<Table.Row class="cursor-pointer hover:bg-accent/50" onclick={() => goto(`/company/${result.company?.symbol}`)}>
								<Table.Cell>
									<div class="flex items-center gap-3">
										<div class="flex h-8 w-8 items-center justify-center rounded-md bg-primary/10 text-xs font-bold text-primary">
											{result.company?.symbol.slice(0, 2)}
										</div>
										<div>
											<div class="font-mono font-medium">{result.company?.symbol}</div>
											<div class="text-xs text-muted-foreground line-clamp-1">{result.company?.name}</div>
										</div>
									</div>
								</Table.Cell>
								<Table.Cell class="text-right font-mono tabular-nums">
									{result.price ? formatPriceCompact(result.price.ltp) : '-'}
								</Table.Cell>
								<Table.Cell class="text-right font-mono tabular-nums {isPositive ? 'text-positive' : isNegative ? 'text-negative' : ''}">
									{result.price ? formatChange(result.price.percentChange) : '-'}
								</Table.Cell>
								<Table.Cell class="hidden text-right font-mono tabular-nums md:table-cell">
									{result.price ? formatVolume(result.price.volume) : '-'}
								</Table.Cell>
								<Table.Cell class="hidden text-right font-mono tabular-nums lg:table-cell">
									{result.fundamentals?.marketCap ? formatMarketCap(result.fundamentals.marketCap) : '-'}
								</Table.Cell>
								<Table.Cell class="hidden text-right font-mono tabular-nums lg:table-cell">
									{result.fundamentals?.pe ? result.fundamentals.pe.toFixed(2) : '-'}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="mt-4 flex items-center justify-between">
					<div class="text-sm text-muted-foreground">
						Page {currentPage} of {totalPages}
					</div>
					<div class="flex gap-2">
						<Button
							variant="outline"
							size="icon"
							disabled={currentPage === 1}
							onclick={() => goto(buildUrl({ offset: (currentPage - 2) * (data.filters.limit ?? 50) }), { invalidateAll: true })}
						>
							<ChevronLeftIcon class="h-4 w-4" />
						</Button>
						<Button
							variant="outline"
							size="icon"
							disabled={currentPage === totalPages}
							onclick={() => goto(buildUrl({ offset: currentPage * (data.filters.limit ?? 50) }), { invalidateAll: true })}
						>
							<ChevronRightIcon class="h-4 w-4" />
						</Button>
					</div>
				</div>
			{/if}
		{:else}
			<div class="rounded-lg border bg-muted/30 py-16 text-center">
				<SlidersHorizontalIcon class="mx-auto h-12 w-12 text-muted-foreground/50" />
				<h3 class="mt-4 font-semibold">No results found</h3>
				<p class="mt-2 text-sm text-muted-foreground">
					Try adjusting your filter criteria
				</p>
			</div>
		{/if}
	</PageContainer>
</section>
