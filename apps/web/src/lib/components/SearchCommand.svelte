<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import type { Company, Price } from '$lib/gen/ntx/v1/common_pb';
	import { Sector } from '$lib/gen/ntx/v1/common_pb';
	import Search from '@lucide/svelte/icons/search';
	import TrendingUp from '@lucide/svelte/icons/trending-up';
	import TrendingDown from '@lucide/svelte/icons/trending-down';
	import X from '@lucide/svelte/icons/x';

	interface Props {
		companies?: Company[];
		prices?: Price[];
		variant?: 'hero' | 'compact';
		placeholder?: string;
		autofocus?: boolean;
	}

	let {
		companies = [],
		prices = [],
		variant = 'compact',
		placeholder = 'Search stocks...',
		autofocus = false
	}: Props = $props();

	let query = $state('');
	let isOpen = $state(false);
	let selectedIndex = $state(0);
	let inputEl = $state<HTMLInputElement | null>(null);

	const sectorLabels: Record<number, string> = {
		[Sector.COMMERCIAL_BANK]: 'Bank',
		[Sector.DEVELOPMENT_BANK]: 'Dev Bank',
		[Sector.FINANCE]: 'Finance',
		[Sector.MICROFINANCE]: 'MFI',
		[Sector.LIFE_INSURANCE]: 'Life Ins.',
		[Sector.NON_LIFE_INSURANCE]: 'Non-Life',
		[Sector.HYDROPOWER]: 'Hydro',
		[Sector.MANUFACTURING]: 'Mfg.',
		[Sector.HOTEL]: 'Hotel',
		[Sector.TRADING]: 'Trading',
		[Sector.INVESTMENT]: 'Invest.',
		[Sector.MUTUAL_FUND]: 'MF',
		[Sector.OTHERS]: 'Other'
	};

	let results = $derived.by(() => {
		if (!query.trim() || companies.length === 0) return [];
		const q = query.toLowerCase();
		return companies
			.filter((c) => c.symbol.toLowerCase().includes(q) || c.name.toLowerCase().includes(q))
			.slice(0, 8)
			.map((c) => ({
				company: c,
				price: prices.find((p) => p.companyId === c.id)
			}));
	});

	function getPrice(p: Price | undefined): number | undefined {
		return p?.ltp ?? p?.close;
	}

	function formatPrice(value: number | undefined): string {
		if (value === undefined) return '—';
		return value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	function handleSelect(symbol: string) {
		goto(`/company/${symbol}`);
		query = '';
		isOpen = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (!isOpen || results.length === 0) return;

		if (e.key === 'ArrowDown') {
			e.preventDefault();
			selectedIndex = (selectedIndex + 1) % results.length;
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			selectedIndex = (selectedIndex - 1 + results.length) % results.length;
		} else if (e.key === 'Enter') {
			e.preventDefault();
			handleSelect(results[selectedIndex].company.symbol);
		} else if (e.key === 'Escape') {
			isOpen = false;
			inputEl?.blur();
		}
	}

	function handleGlobalKeydown(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
			e.preventDefault();
			inputEl?.focus();
		}
	}

	$effect(() => {
		window.addEventListener('keydown', handleGlobalKeydown);
		return () => window.removeEventListener('keydown', handleGlobalKeydown);
	});

	$effect(() => {
		if (query.trim()) {
			isOpen = true;
			selectedIndex = 0;
		} else {
			isOpen = false;
		}
	});
</script>

<div class="relative w-full" class:max-w-xl={variant === 'hero'} class:max-w-xs={variant === 'compact'}>
	{#if variant === 'hero'}
		<!-- Hero variant: larger, centered search -->
		<div class="group relative">
			<div
				class="absolute -inset-1 rounded-2xl bg-gradient-to-r from-chart-1/20 via-positive/10 to-chart-2/20 opacity-0 blur-xl transition-opacity duration-500 group-focus-within:opacity-100"
			></div>
			<div class="relative flex items-center rounded-xl border border-border bg-card shadow-lg">
				<Search class="ml-5 size-5 text-muted-foreground" />
				<input
					bind:this={inputEl}
					type="text"
					{placeholder}
					bind:value={query}
					onkeydown={handleKeydown}
					onfocus={() => query.trim() && (isOpen = true)}
					onblur={() => setTimeout(() => (isOpen = false), 200)}
					class="flex-1 border-none bg-transparent px-4 py-4 text-lg placeholder:text-muted-foreground focus:ring-0 focus:outline-none"
				/>
				<kbd class="mr-4 hidden rounded border border-border bg-muted px-2 py-0.5 text-xs text-muted-foreground sm:inline">
					⌘K
				</kbd>
			</div>
		</div>
	{:else}
		<!-- Compact variant: navbar style -->
		<div class="relative flex items-center rounded-lg border border-border bg-background">
			<Search class="ml-3 size-4 text-muted-foreground" />
			<input
				bind:this={inputEl}
				type="text"
				{placeholder}
				bind:value={query}
				onkeydown={handleKeydown}
				onfocus={() => query.trim() && (isOpen = true)}
				onblur={() => setTimeout(() => (isOpen = false), 200)}
				class="w-full border-none bg-transparent px-3 py-2 text-sm placeholder:text-muted-foreground focus:ring-0 focus:outline-none"
			/>
			{#if query}
				<button
					onclick={() => (query = '')}
					class="mr-2 rounded p-0.5 text-muted-foreground hover:text-foreground"
				>
					<X class="size-3" />
				</button>
			{:else}
				<kbd class="mr-2 hidden rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground lg:inline">
					⌘K
				</kbd>
			{/if}
		</div>
	{/if}

	<!-- Results dropdown -->
	{#if isOpen && results.length > 0}
		<div
			class="absolute top-full z-50 mt-2 w-full overflow-hidden rounded-xl border border-border bg-card shadow-2xl"
			class:min-w-[400px]={variant === 'compact'}
		>
			<div class="max-h-[400px] overflow-y-auto">
				{#each results as result, i (result.company.id)}
					{@const currentPrice = getPrice(result.price)}
					{@const change = result.price?.changePercent}
					<button
						onclick={() => handleSelect(result.company.symbol)}
						class="flex w-full items-center justify-between px-4 py-3 text-left transition-colors
							{selectedIndex === i ? 'bg-muted' : 'hover:bg-muted/50'}"
					>
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-2">
								<span class="font-serif text-lg font-medium">{result.company.symbol}</span>
								<span class="rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
									{sectorLabels[result.company.sector ?? Sector.OTHERS]}
								</span>
							</div>
							<p class="truncate text-sm text-muted-foreground">{result.company.name}</p>
						</div>
						<div class="ml-4 text-right">
							<div class="tabular-nums">Rs. {formatPrice(currentPrice)}</div>
							{#if change !== undefined}
								<div
									class="flex items-center justify-end gap-1 text-sm tabular-nums
										{change >= 0 ? 'text-positive' : 'text-negative'}"
								>
									{#if change >= 0}
										<TrendingUp class="size-3" />
									{:else}
										<TrendingDown class="size-3" />
									{/if}
									{change > 0 ? '+' : ''}{change.toFixed(2)}%
								</div>
							{/if}
						</div>
					</button>
				{/each}
			</div>
			<div class="border-t border-border bg-muted/30 px-4 py-2">
				<p class="text-xs text-muted-foreground">
					<kbd class="rounded border border-border bg-background px-1">↑↓</kbd> navigate
					<kbd class="ml-2 rounded border border-border bg-background px-1">↵</kbd> select
					<kbd class="ml-2 rounded border border-border bg-background px-1">esc</kbd> close
				</p>
			</div>
		</div>
	{/if}
</div>
