<script lang="ts">
	import CompanyCard from '$lib/components/CompanyCard.svelte';
	import SearchCommand from '$lib/components/SearchCommand.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { Sector } from '$lib/gen/ntx/v1/common_pb';
	import type { Company, Price } from '$lib/gen/ntx/v1/common_pb';
	import TrendingUp from '@lucide/svelte/icons/trending-up';
	import TrendingDown from '@lucide/svelte/icons/trending-down';

	let { data } = $props();

	let selectedSector = $state<number | null>(null);

	const sectors = [
		{ value: Sector.COMMERCIAL_BANK, label: 'Banks', short: 'BNK' },
		{ value: Sector.DEVELOPMENT_BANK, label: 'Dev Banks', short: 'DEV' },
		{ value: Sector.FINANCE, label: 'Finance', short: 'FIN' },
		{ value: Sector.MICROFINANCE, label: 'Microfinance', short: 'MFI' },
		{ value: Sector.LIFE_INSURANCE, label: 'Life Insurance', short: 'LIF' },
		{ value: Sector.NON_LIFE_INSURANCE, label: 'Non-Life', short: 'NLI' },
		{ value: Sector.HYDROPOWER, label: 'Hydropower', short: 'HYD' },
		{ value: Sector.MANUFACTURING, label: 'Manufacturing', short: 'MFG' },
		{ value: Sector.HOTEL, label: 'Hotels', short: 'HTL' },
		{ value: Sector.TRADING, label: 'Trading', short: 'TRD' },
		{ value: Sector.INVESTMENT, label: 'Investment', short: 'INV' },
		{ value: Sector.OTHERS, label: 'Others', short: 'OTH' }
	];

	// Get price for a company
	function getPrice(companyId: bigint): Price | undefined {
		return data.prices?.find((p) => p.companyId === companyId);
	}

	function formatPrice(value: number | undefined): string {
		if (value === undefined) return '—';
		return value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	function formatChange(value: number | undefined): string {
		if (value === undefined) return '';
		return `${value > 0 ? '+' : ''}${value.toFixed(2)}%`;
	}

	// Top gainers
	let topGainers = $derived.by(() => {
		if (!data.prices) return [];
		return [...data.prices]
			.filter((p) => p.changePercent !== undefined && p.changePercent > 0)
			.sort((a, b) => (b.changePercent ?? 0) - (a.changePercent ?? 0))
			.slice(0, 5)
			.map((p) => ({
				price: p,
				company: data.companies?.find((c) => c.id === p.companyId)
			}))
			.filter((x) => x.company);
	});

	// Top losers
	let topLosers = $derived.by(() => {
		if (!data.prices) return [];
		return [...data.prices]
			.filter((p) => p.changePercent !== undefined && p.changePercent < 0)
			.sort((a, b) => (a.changePercent ?? 0) - (b.changePercent ?? 0))
			.slice(0, 5)
			.map((p) => ({
				price: p,
				company: data.companies?.find((c) => c.id === p.companyId)
			}))
			.filter((x) => x.company);
	});

	// Most traded by volume
	let mostTraded = $derived.by(() => {
		if (!data.prices) return [];
		return [...data.prices]
			.filter((p) => p.volume !== undefined && p.volume > 0)
			.sort((a, b) => Number(b.volume ?? 0) - Number(a.volume ?? 0))
			.slice(0, 5)
			.map((p) => ({
				price: p,
				company: data.companies?.find((c) => c.id === p.companyId)
			}))
			.filter((x) => x.company);
	});

	// Filtered companies for sector view
	let filteredCompanies = $derived.by(() => {
		if (selectedSector === null) return [];
		return (data.companies ?? []).filter((c) => c.sector === selectedSector);
	});

	// Sector stats
	let sectorCounts = $derived.by(() => {
		const counts: Record<number, number> = {};
		for (const c of data.companies ?? []) {
			counts[c.sector ?? 0] = (counts[c.sector ?? 0] ?? 0) + 1;
		}
		return counts;
	});

	function toggleSector(sector: number) {
		selectedSector = selectedSector === sector ? null : sector;
	}

	// Get today's date
	const today = new Date().toLocaleDateString('en-US', {
		weekday: 'long',
		year: 'numeric',
		month: 'long',
		day: 'numeric'
	});

	// Random companies for discovery section
	let randomCompanies = $derived.by(() => {
		const companies = data.companies ?? [];
		if (companies.length === 0) return [];

		// Fisher-Yates shuffle
		const shuffled = [...companies];
		for (let i = shuffled.length - 1; i > 0; i--) {
			const j = Math.floor(Math.random() * (i + 1));
			[shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
		}

		return shuffled.slice(0, 6);
	});
</script>

<div class="min-h-screen">
	<!-- Masthead -->
	<header class="border-b border-border">
		<div class="mx-auto max-w-7xl px-4">
			<!-- Top bar -->
			<div class="flex items-center justify-between py-3 text-xs text-muted-foreground">
				<span>{today}</span>
				<div class="flex items-center gap-3">
					<span>{data.companies?.length ?? 0} Securities Listed</span>
					<ThemeToggle />
				</div>
			</div>

			<!-- Logo & Search -->
			<div
				class="flex flex-col items-center gap-6 border-t border-border py-8 md:flex-row md:justify-between"
			>
				<a href="/" class="text-center md:text-left">
					<h1 class="font-serif text-4xl tracking-tight md:text-5xl">NTX</h1>
					<p class="text-sm tracking-[0.3em] text-muted-foreground uppercase">Market Stories</p>
				</a>

				<SearchCommand
					companies={data.companies ?? []}
					prices={data.prices ?? []}
					variant="hero"
					placeholder="Search {data.companies?.length ?? 0} stocks..."
				/>
			</div>
		</div>
	</header>

	<!-- Main Content -->
	<main class="mx-auto max-w-7xl px-4 py-8">
		<div class="grid gap-8 lg:grid-cols-[1fr_300px]">
			<!-- Market Movers: First on mobile, second on desktop -->
			<aside class="space-y-6 overflow-hidden lg:order-2">
				<!-- Top Gainers -->
				<div class="rounded-xl border border-border bg-card p-4">
					<div class="mb-3 flex items-center gap-2">
						<TrendingUp class="size-4 text-positive" />
						<h3 class="font-medium">Top Gainers</h3>
					</div>
					<div class="space-y-1">
						{#each topGainers as item (item.company?.id)}
							<a
								href="/company/{item.company?.symbol}"
								class="flex items-center justify-between gap-2 rounded-lg px-2 py-2 transition-colors hover:bg-muted"
							>
								<div class="min-w-0 flex-1">
									<span class="font-serif">{item.company?.symbol}</span>
									<p class="truncate text-xs text-muted-foreground">{item.company?.name}</p>
								</div>
								<span class="shrink-0 font-medium text-positive tabular-nums">
									{formatChange(item.price.changePercent)}
								</span>
							</a>
						{/each}
					</div>
				</div>

				<!-- Top Losers -->
				<div class="rounded-xl border border-border bg-card p-4">
					<div class="mb-3 flex items-center gap-2">
						<TrendingDown class="size-4 text-negative" />
						<h3 class="font-medium">Top Losers</h3>
					</div>
					<div class="space-y-1">
						{#each topLosers as item (item.company?.id)}
							<a
								href="/company/{item.company?.symbol}"
								class="flex items-center justify-between gap-2 rounded-lg px-2 py-2 transition-colors hover:bg-muted"
							>
								<div class="min-w-0 flex-1">
									<span class="font-serif">{item.company?.symbol}</span>
									<p class="truncate text-xs text-muted-foreground">{item.company?.name}</p>
								</div>
								<span class="shrink-0 font-medium text-negative tabular-nums">
									{formatChange(item.price.changePercent)}
								</span>
							</a>
						{/each}
					</div>
				</div>

				<!-- Most Traded -->
				<div class="rounded-xl border border-border bg-card p-4">
					<div class="mb-3 flex items-center gap-2">
						<div class="size-2 rounded-full bg-chart-1"></div>
						<h3 class="font-medium">Most Traded</h3>
					</div>
					<div class="space-y-1">
						{#each mostTraded as item (item.company?.id)}
							<a
								href="/company/{item.company?.symbol}"
								class="flex items-center justify-between gap-2 rounded-lg px-2 py-2 transition-colors hover:bg-muted"
							>
								<div class="min-w-0 flex-1">
									<span class="font-serif">{item.company?.symbol}</span>
									<p class="truncate text-xs text-muted-foreground">{item.company?.name}</p>
								</div>
								<span class="shrink-0 text-sm text-muted-foreground tabular-nums">
									{Number(item.price.volume).toLocaleString()}
								</span>
							</a>
						{/each}
					</div>
				</div>
			</aside>

			<!-- Sector Explorer: Second on mobile, first on desktop -->
			<div class="lg:order-1">
				<div class="mb-6 flex items-end justify-between">
					<div>
						<h2 class="font-serif text-2xl">Explore by Sector</h2>
						<p class="mt-1 text-sm text-muted-foreground">
							{selectedSector !== null
								? `${filteredCompanies.length} companies in sector`
								: 'Select a sector to browse'}
						</p>
					</div>
					{#if selectedSector !== null}
						<button
							onclick={() => (selectedSector = null)}
							class="text-sm text-muted-foreground hover:text-foreground hover:underline"
						>
							Clear selection
						</button>
					{/if}
				</div>

				<!-- Sector Grid -->
				<div class="mb-8 grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-6">
					{#each sectors as sector (sector.value)}
						{@const count = sectorCounts[sector.value] ?? 0}
						<button
							onclick={() => toggleSector(sector.value)}
							class="group relative overflow-hidden rounded-lg border p-3 text-left transition-all
								{selectedSector === sector.value
								? 'border-primary bg-primary/10 ring-1 ring-primary/20'
								: 'border-border hover:border-foreground/50'}"
						>
							<span class="text-2xl font-medium tabular-nums opacity-20">{count}</span>
							<p class="mt-1 text-xs font-medium">{sector.label}</p>
						</button>
					{/each}
				</div>

				<!-- Company Grid -->
				{#if selectedSector !== null}
					<div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
						{#each filteredCompanies as company (company.id)}
							{@const price = getPrice(company.id)}
							<CompanyCard {company} {price} />
						{/each}
					</div>
				{:else}
					<!-- Random featured companies when no sector selected -->
					<div>
						<h3 class="mb-4 font-serif text-xl">Discover Companies</h3>
						<div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
							{#each randomCompanies as company (company.id)}
								{@const price = getPrice(company.id)}
								<CompanyCard {company} {price} />
							{/each}
						</div>
					</div>
				{/if}
			</div>
		</div>
	</main>
</div>
