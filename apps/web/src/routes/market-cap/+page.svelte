<script lang="ts">
	import type { Price } from '$lib/gen/ntx/v1/common_pb';
	import MarketCapTreemap from '$lib/components/charts/MarketCapTreemap.svelte';
	import LayoutGrid from '@lucide/svelte/icons/layout-grid';
	import List from '@lucide/svelte/icons/list';

	let { data } = $props();

	let viewMode = $state<'treemap' | 'table'>('treemap');

	function getPrice(companyId: bigint): Price | undefined {
		return data.prices?.find((p) => p.companyId === companyId);
	}

	function formatLarge(value: number): string {
		if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`;
		if (value >= 10_000_000) return `${(value / 10_000_000).toFixed(2)} Cr`;
		if (value >= 100_000) return `${(value / 100_000).toFixed(2)} L`;
		return value.toLocaleString();
	}

	function formatCurrency(value: number) {
		return new Intl.NumberFormat('en-NP', {
			style: 'currency',
			currency: 'NPR',
			maximumFractionDigits: 2
		}).format(value);
	}

	let rankings = $derived.by(() => {
		if (!data.companies) return [];
		return [...data.companies]
			.map((c) => {
				const price = getPrice(c.id);
				const shares = c.listedShares ? Number(c.listedShares) : 0;
				const ltp = price?.ltp ?? price?.close ?? 0;
				return {
					company: c,
					price,
					marketCap: shares * ltp,
					shares,
					ltp
				};
			})
			.filter((x) => x.marketCap > 0)
			.sort((a, b) => b.marketCap - a.marketCap);
	});

	// Calculate total market cap
	let totalMarketCap = $derived(rankings.reduce((sum, r) => sum + r.marketCap, 0));
</script>

<div class="min-h-screen bg-background">
	<div class="mx-auto max-w-7xl px-4 py-8">
		<!-- Header -->
		<div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
			<div>
				<h1 class="font-serif text-2xl font-medium">Market Capitalization</h1>
				<p class="mt-1 text-sm text-muted-foreground">
					{rankings.length} companies · Total: {formatLarge(totalMarketCap)}
				</p>
			</div>

			<!-- View Toggle -->
			<div class="flex items-center gap-1 rounded-lg border border-border bg-muted/30 p-1">
				<button
					onclick={() => (viewMode = 'treemap')}
					class="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm transition-colors {viewMode ===
					'treemap'
						? 'bg-background text-foreground shadow-sm'
						: 'text-muted-foreground hover:text-foreground'}"
				>
					<LayoutGrid class="size-4" />
					<span class="hidden sm:inline">Treemap</span>
				</button>
				<button
					onclick={() => (viewMode = 'table')}
					class="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm transition-colors {viewMode ===
					'table'
						? 'bg-background text-foreground shadow-sm'
						: 'text-muted-foreground hover:text-foreground'}"
				>
					<List class="size-4" />
					<span class="hidden sm:inline">Table</span>
				</button>
			</div>
		</div>

		<!-- Treemap View -->
		{#if viewMode === 'treemap'}
			<div class="rounded-xl border border-border bg-card p-4">
				<MarketCapTreemap
					companies={data.companies ?? []}
					prices={data.prices ?? []}
					class="h-[500px] sm:h-[600px]"
				/>
			</div>
		{:else}
			<!-- Table View -->
			<div class="overflow-hidden rounded-xl border border-border bg-card">
				<div class="overflow-x-auto">
					<table class="w-full text-left text-sm">
						<thead class="bg-muted/50 text-xs uppercase text-muted-foreground">
							<tr>
								<th class="w-12 px-4 py-3 font-medium">Rank</th>
								<th class="px-4 py-3 font-medium">Company</th>
								<th class="px-4 py-3 text-right font-medium">Price</th>
								<th class="px-4 py-3 text-right font-medium">Listed Shares</th>
								<th class="px-4 py-3 text-right font-medium">Market Cap</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-border">
							{#each rankings as item, i (item.company.id)}
								<tr class="transition-colors hover:bg-muted/50">
									<td class="w-12 px-4 py-3 font-medium tabular-nums text-muted-foreground">
										{i + 1}
									</td>
									<td class="px-4 py-3">
										<a
											href="/company/{item.company.symbol}"
											class="group flex items-center gap-3"
										>
											<div>
												<div class="font-medium transition-colors group-hover:text-primary">
													{item.company.symbol}
												</div>
												<div
													class="max-w-[200px] truncate text-xs text-muted-foreground sm:max-w-none"
												>
													{item.company.name}
												</div>
											</div>
										</a>
									</td>
									<td class="px-4 py-3 text-right tabular-nums">
										{formatCurrency(item.ltp)}
									</td>
									<td class="px-4 py-3 text-right tabular-nums text-muted-foreground">
										{formatLarge(item.shares)}
									</td>
									<td class="px-4 py-3 text-right font-medium tabular-nums">
										{formatLarge(item.marketCap)}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		{/if}
	</div>
</div>
