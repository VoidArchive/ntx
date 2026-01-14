<script lang="ts">
	import { Chart, Svg, Treemap, Group, Rect, Text } from 'layerchart';
	import { hierarchy, type HierarchyRectangularNode } from 'd3-hierarchy';
	import { Sector } from '$lib/gen/ntx/v1/common_pb';
	import type { Company, Price } from '$lib/gen/ntx/v1/common_pb';

	interface Props {
		companies: Company[];
		prices: Price[];
		class?: string;
	}

	let { companies, prices, class: className = '' }: Props = $props();

	// Data node type
	interface TreeNode {
		name: string;
		symbol?: string;
		sector?: number;
		value?: number;
		volume?: number;
		children?: TreeNode[];
	}

	// Sector colors (same as MarketCapTreemap)
	const sectorColors: Record<number, string> = {
		[Sector.COMMERCIAL_BANK]: 'oklch(0.55 0.12 200)',
		[Sector.DEVELOPMENT_BANK]: 'oklch(0.58 0.14 155)',
		[Sector.FINANCE]: 'oklch(0.62 0.12 45)',
		[Sector.MICROFINANCE]: 'oklch(0.55 0.1 280)',
		[Sector.LIFE_INSURANCE]: 'oklch(0.65 0.12 30)',
		[Sector.NON_LIFE_INSURANCE]: 'oklch(0.5 0.15 340)',
		[Sector.HYDROPOWER]: 'oklch(0.6 0.18 200)',
		[Sector.MANUFACTURING]: 'oklch(0.55 0.1 90)',
		[Sector.HOTEL]: 'oklch(0.6 0.12 60)',
		[Sector.TRADING]: 'oklch(0.5 0.12 180)',
		[Sector.INVESTMENT]: 'oklch(0.55 0.15 300)',
		[Sector.OTHERS]: 'oklch(0.5 0.05 30)'
	};

	const sectorLabels: Record<number, string> = {
		[Sector.COMMERCIAL_BANK]: 'Banks',
		[Sector.DEVELOPMENT_BANK]: 'Dev Banks',
		[Sector.FINANCE]: 'Finance',
		[Sector.MICROFINANCE]: 'Microfinance',
		[Sector.LIFE_INSURANCE]: 'Life Insurance',
		[Sector.NON_LIFE_INSURANCE]: 'Non-Life Insurance',
		[Sector.HYDROPOWER]: 'Hydropower',
		[Sector.MANUFACTURING]: 'Manufacturing',
		[Sector.HOTEL]: 'Hotels',
		[Sector.TRADING]: 'Trading',
		[Sector.INVESTMENT]: 'Investment',
		[Sector.OTHERS]: 'Others'
	};

	function getCompany(companyId: bigint): Company | undefined {
		return companies?.find((c) => c.id === companyId);
	}

	function formatVolume(value: number): string {
		if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
		if (value >= 1_000) return `${(value / 1_000).toFixed(1)}K`;
		return value.toLocaleString();
	}

	// Build flat data for Chart
	let flatData = $derived.by(() => {
		if (!prices || prices.length === 0) return [];

		return prices
			.map((price) => {
				const company = getCompany(price.companyId);
				if (!company) return null;
				const volume = Number(price.volume ?? 0);
				return {
					name: company.name ?? '',
					symbol: company.symbol ?? '',
					sector: company.sector ?? Sector.OTHERS,
					value: volume,
					volume
				};
			})
			.filter((x): x is NonNullable<typeof x> => x !== null && x.value > 0);
	});

	// Build hierarchy data grouped by sector
	let hierarchyRoot = $derived.by(() => {
		if (flatData.length === 0) return null;

		const sectorGroups: Record<number, TreeNode[]> = {};

		for (const item of flatData) {
			const sector = item.sector;
			if (!sectorGroups[sector]) {
				sectorGroups[sector] = [];
			}
			sectorGroups[sector].push({
				name: item.name,
				symbol: item.symbol,
				value: item.value,
				volume: item.volume
			});
		}

		const children: TreeNode[] = Object.entries(sectorGroups)
			.map(([sectorKey, comps]) => ({
				name: sectorLabels[Number(sectorKey)] ?? 'Others',
				sector: Number(sectorKey),
				children: comps.sort((a, b) => (b.value ?? 0) - (a.value ?? 0))
			}))
			.filter((s) => s.children && s.children.length > 0)
			.sort((a, b) => {
				const aTotal = a.children?.reduce((sum, c) => sum + (c.value ?? 0), 0) ?? 0;
				const bTotal = b.children?.reduce((sum, c) => sum + (c.value ?? 0), 0) ?? 0;
				return bTotal - aTotal;
			});

		const root: TreeNode = { name: 'root', children };
		return hierarchy(root).sum((d: TreeNode) => d.value ?? 0);
	});

	// Hover state
	let hoveredNode = $state<{
		symbol: string;
		name: string;
		volume: number;
		sector: string;
		x: number;
		y: number;
	} | null>(null);

	function handleMouseEnter(node: HierarchyRectangularNode<TreeNode>, event: MouseEvent) {
		if (node.depth !== 2) return;

		const rect = (event.currentTarget as SVGElement).getBoundingClientRect();
		const container = (event.currentTarget as SVGElement).closest('svg')?.getBoundingClientRect();

		if (!container) return;

		hoveredNode = {
			symbol: node.data.symbol ?? '',
			name: node.data.name,
			volume: node.data.volume ?? 0,
			sector: sectorLabels[node.parent?.data?.sector ?? Sector.OTHERS] ?? 'Other',
			x: rect.left - container.left + rect.width / 2,
			y: rect.top - container.top
		};
	}

	function handleMouseLeave() {
		hoveredNode = null;
	}

	function handleClick(node: HierarchyRectangularNode<TreeNode>) {
		if (node.depth === 2 && node.data.symbol) {
			window.location.href = `/company/${node.data.symbol}`;
		}
	}
</script>

{#if hierarchyRoot && flatData.length > 0}
	<div class="relative {className}">
		<Chart data={flatData} padding={{ top: 0, bottom: 0, left: 0, right: 0 }}>
			<Svg>
				<Treemap hierarchy={hierarchyRoot} padding={2}>
					{#snippet children({ nodes })}
						{#each nodes as node (node.data.name + node.depth)}
							{#if node.depth === 1}
								<Rect
									x={node.x0}
									y={node.y0}
									width={node.x1 - node.x0}
									height={node.y1 - node.y0}
									fill={sectorColors[node.data.sector ?? Sector.OTHERS]}
									class="opacity-20"
								/>
							{:else if node.depth === 2}
								{@const width = node.x1 - node.x0}
								{@const height = node.y1 - node.y0}
								{@const showSymbol = width > 40 && height > 25}
								{@const showValue = width > 60 && height > 40}
								<Group>
									<Rect
										x={node.x0}
										y={node.y0}
										{width}
										{height}
										fill={sectorColors[node.parent?.data?.sector ?? Sector.OTHERS]}
										stroke="var(--background)"
										strokeWidth={1}
										class="cursor-pointer transition-opacity hover:opacity-80"
										role="button"
										tabindex={0}
										onmouseenter={(e) => handleMouseEnter(node, e)}
										onmouseleave={handleMouseLeave}
										onclick={() => handleClick(node)}
										onkeypress={(e) => e.key === 'Enter' && handleClick(node)}
									/>
									{#if showSymbol}
										<Text
											x={(node.x0 + node.x1) / 2}
											y={(node.y0 + node.y1) / 2 - (showValue ? 6 : 0)}
											textAnchor="middle"
											verticalAnchor="middle"
											value={node.data.symbol ?? ''}
											class="pointer-events-none fill-white text-[10px] font-medium drop-shadow-sm"
										/>
									{/if}
									{#if showValue}
										<Text
											x={(node.x0 + node.x1) / 2}
											y={(node.y0 + node.y1) / 2 + 8}
											textAnchor="middle"
											verticalAnchor="middle"
											value={formatVolume(node.data.value ?? 0)}
											class="pointer-events-none fill-white/80 text-[9px] drop-shadow-sm"
										/>
									{/if}
								</Group>
							{/if}
						{/each}
					{/snippet}
				</Treemap>
			</Svg>
		</Chart>

		<!-- Tooltip -->
		{#if hoveredNode}
			<div
				class="pointer-events-none absolute z-10 rounded-lg border border-border bg-popover px-3 py-2 text-xs shadow-lg"
				style="left: {hoveredNode.x}px; top: {hoveredNode.y}px; transform: translate(-50%, -100%) translateY(-8px);"
			>
				<p class="font-medium">{hoveredNode.symbol}</p>
				<p class="mt-0.5 max-w-[200px] truncate text-muted-foreground">{hoveredNode.name}</p>
				<div class="mt-1 flex items-center justify-between gap-4">
					<span class="text-muted-foreground">{hoveredNode.sector}</span>
					<span class="font-medium tabular-nums">{formatVolume(hoveredNode.volume)}</span>
				</div>
			</div>
		{/if}

		<!-- Legend -->
		<div class="mt-4 flex flex-wrap gap-3 text-xs">
			{#each Object.entries(sectorLabels) as [sectorKey, label] (sectorKey)}
				{@const sector = Number(sectorKey)}
				{@const color = sectorColors[sector]}
				<div class="flex items-center gap-1.5">
					<div class="size-3 rounded-sm" style="background: {color};"></div>
					<span class="text-muted-foreground">{label}</span>
				</div>
			{/each}
		</div>
	</div>
{:else}
	<div class="flex h-[400px] items-center justify-center rounded-xl border border-border bg-muted/30">
		<p class="text-sm text-muted-foreground">No volume data available</p>
	</div>
{/if}
