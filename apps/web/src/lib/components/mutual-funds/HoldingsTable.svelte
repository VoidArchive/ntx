<script lang="ts">
	import type { Holdings, Holding } from '$lib/types/fund';
	import { SECTOR_LABELS, SECTOR_COLORS } from '$lib/types/fund';
	import { SvelteSet } from 'svelte/reactivity';
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';

	interface Props {
		holdings: Holdings;
		netAssets: number;
	}

	let { holdings, netAssets }: Props = $props();

	// Track expanded sectors
	let expandedSectors = new SvelteSet<string>();

	interface SectorGroup {
		key: keyof Holdings;
		label: string;
		color: string;
		items: Holding[];
		total: number;
		percent: number;
	}

	// Group holdings by sector
	let sectorGroups = $derived.by((): SectorGroup[] => {
		const groups: SectorGroup[] = [];

		for (const [key, items] of Object.entries(holdings)) {
			if (!Array.isArray(items) || items.length === 0) continue;

			const total = items.reduce((sum, item) => sum + item.value, 0);
			groups.push({
				key: key as keyof Holdings,
				label: SECTOR_LABELS[key as keyof Holdings] || key,
				color: SECTOR_COLORS[key as keyof Holdings] || '#64748b',
				items: [...items].sort((a, b) => b.value - a.value),
				total,
				percent: (total / netAssets) * 100
			});
		}

		return groups.sort((a, b) => b.total - a.total);
	});

	function toggleSector(key: string) {
		if (expandedSectors.has(key)) {
			expandedSectors.delete(key);
		} else {
			expandedSectors.add(key);
		}
	}

	function fmtValue(value: number): string {
		if (value >= 10_000_000) return `${(value / 10_000_000).toFixed(2)} Cr`;
		if (value >= 100_000) return `${(value / 100_000).toFixed(2)} L`;
		return value.toLocaleString('en-NP');
	}

	function fmtUnits(units: number | undefined): string {
		if (!units) return '-';
		return units.toLocaleString('en-NP');
	}
</script>

<div>
	<h3 class="mb-4 font-serif text-lg font-medium">Holdings by Sector</h3>

	<div class="space-y-2">
		{#each sectorGroups as group (group.key)}
			{@const isExpanded = expandedSectors.has(group.key)}
			<div class="overflow-hidden rounded-lg border border-border">
				<!-- Sector Header -->
				<button
					class="flex w-full items-center gap-2 bg-card/50 px-3 py-3 text-left transition-colors hover:bg-muted/50 sm:gap-3 sm:px-4"
					onclick={() => toggleSector(group.key)}
				>
					{#if isExpanded}
						<ChevronDown class="size-4 shrink-0 text-muted-foreground" />
					{:else}
						<ChevronRight class="size-4 shrink-0 text-muted-foreground" />
					{/if}

					<div class="size-3 shrink-0 rounded-sm" style="background: {group.color};"></div>

					<span class="min-w-0 flex-1 truncate text-sm font-medium sm:text-base">{group.label}</span
					>

					<span class="hidden text-sm text-muted-foreground tabular-nums sm:inline">
						{group.items.length}
						{group.items.length === 1 ? 'holding' : 'holdings'}
					</span>

					<span class="w-14 text-right text-sm font-medium tabular-nums sm:w-20">
						{group.percent.toFixed(1)}%
					</span>

					<span class="hidden w-24 text-right text-sm text-muted-foreground tabular-nums sm:inline">
						{fmtValue(group.total)}
					</span>
				</button>

				<!-- Holdings List -->
				{#if isExpanded}
					<div class="border-t border-border bg-background">
						<!-- Mobile: Card layout -->
						<div class="sm:hidden">
							{#each group.items as item (item.name)}
								<div class="flex items-start justify-between gap-2 px-3 py-2">
									<span class="text-sm">{item.name}</span>
									<span class="shrink-0 text-sm text-muted-foreground tabular-nums">
										{fmtValue(item.value)}
									</span>
								</div>
							{/each}
						</div>

						<!-- Desktop: Table layout -->
						<table class="hidden w-full text-sm sm:table">
							<thead>
								<tr class="border-b border-border text-xs text-muted-foreground">
									<th class="px-4 py-2 text-left font-medium">Company</th>
									<th class="px-4 py-2 text-right font-medium">Units</th>
									<th class="px-4 py-2 text-right font-medium">Value</th>
									<th class="w-20 px-4 py-2 text-right font-medium">%</th>
								</tr>
							</thead>
							<tbody>
								{#each group.items as item (item.name)}
									<tr class="hover:bg-muted/30">
										<td class="px-4 py-2">{item.name}</td>
										<td class="px-4 py-2 text-right text-muted-foreground tabular-nums">
											{fmtUnits(item.units)}
										</td>
										<td class="px-4 py-2 text-right tabular-nums">
											{fmtValue(item.value)}
										</td>
										<td class="px-4 py-2 text-right text-muted-foreground tabular-nums">
											{((item.value / netAssets) * 100).toFixed(2)}%
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		{/each}
	</div>
</div>
