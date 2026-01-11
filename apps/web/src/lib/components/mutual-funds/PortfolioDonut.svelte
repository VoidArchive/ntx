<script lang="ts">
	import type { Holdings } from '$lib/types/fund';
	import { SECTOR_LABELS, SECTOR_COLORS } from '$lib/types/fund';

	interface Props {
		holdings: Holdings;
		netAssets: number;
	}

	let { holdings, netAssets }: Props = $props();

	const size = 220;
	const center = size / 2;
	const radius = 85;
	const innerRadius = 55;

	interface SectorData {
		key: keyof Holdings;
		label: string;
		value: number;
		percent: number;
		color: string;
	}

	// Calculate sector totals
	let sectorData = $derived.by((): SectorData[] => {
		const data: SectorData[] = [];

		for (const [key, items] of Object.entries(holdings)) {
			if (!Array.isArray(items) || items.length === 0) continue;

			const total = items.reduce((sum, item) => sum + item.value, 0);
			if (total > 0) {
				data.push({
					key: key as keyof Holdings,
					label: SECTOR_LABELS[key as keyof Holdings] || key,
					value: total,
					percent: (total / netAssets) * 100,
					color: SECTOR_COLORS[key as keyof Holdings] || '#64748b'
				});
			}
		}

		// Sort by value descending
		return data.sort((a, b) => b.value - a.value);
	});

	// Hover state
	let hoveredSector = $state<string | null>(null);

	// Create arc paths
	function polarToCartesian(cx: number, cy: number, r: number, angleInDegrees: number) {
		const rad = (angleInDegrees * Math.PI) / 180;
		return { x: cx + r * Math.cos(rad), y: cy + r * Math.sin(rad) };
	}

	function createArc(startPercent: number, endPercent: number): string {
		const startAngle = (startPercent / 100) * 360 - 90;
		const endAngle = (endPercent / 100) * 360 - 90;

		const outerStart = polarToCartesian(center, center, radius, endAngle);
		const outerEnd = polarToCartesian(center, center, radius, startAngle);
		const innerStart = polarToCartesian(center, center, innerRadius, endAngle);
		const innerEnd = polarToCartesian(center, center, innerRadius, startAngle);

		const sweep = endPercent - startPercent;
		const largeArc = sweep > 50 ? 1 : 0;

		return [
			'M',
			outerStart.x,
			outerStart.y,
			'A',
			radius,
			radius,
			0,
			largeArc,
			0,
			outerEnd.x,
			outerEnd.y,
			'L',
			innerEnd.x,
			innerEnd.y,
			'A',
			innerRadius,
			innerRadius,
			0,
			largeArc,
			1,
			innerStart.x,
			innerStart.y,
			'Z'
		].join(' ');
	}

	// Generate arc data with cumulative percentages
	let arcs = $derived.by(() => {
		let cumulative = 0;
		return sectorData.map((sector) => {
			const start = cumulative;
			cumulative += sector.percent;
			return {
				...sector,
				path: createArc(start, cumulative),
				start,
				end: cumulative
			};
		});
	});

	function fmtLarge(value: number): string {
		if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(1)}B`;
		if (value >= 10_000_000) return `${(value / 10_000_000).toFixed(1)} Cr`;
		if (value >= 100_000) return `${(value / 100_000).toFixed(1)} L`;
		return value.toLocaleString('en-NP');
	}

	// Tooltip
	let tooltipX = $state(0);
	let tooltipY = $state(0);

	function handleMouseMove(e: MouseEvent) {
		const rect = (e.currentTarget as SVGElement).getBoundingClientRect();
		tooltipX = e.clientX - rect.left;
		tooltipY = e.clientY - rect.top - 10;
	}
</script>

<div>
	<h3 class="mb-4 font-serif text-lg font-medium">Portfolio Allocation</h3>

	<div class="flex flex-col items-center gap-6 lg:flex-row lg:items-start">
		<!-- Donut Chart -->
		<div class="relative shrink-0">
			<svg
				width={size}
				height={size}
				viewBox="0 0 {size} {size}"
				class="drop-shadow-md"
				role="img"
				aria-label="Portfolio allocation pie chart"
				onmousemove={handleMouseMove}
			>
				<defs>
					<filter id="donut-glow" x="-20%" y="-20%" width="140%" height="140%">
						<feDropShadow dx="0" dy="2" stdDeviation="2" flood-opacity="0.15" />
					</filter>
				</defs>

				<g filter="url(#donut-glow)">
					{#each arcs as arc (arc.key)}
						<path
							d={arc.path}
							fill={arc.color}
							class="cursor-pointer transition-all duration-150"
							style="opacity: {hoveredSector && hoveredSector !== arc.key ? 0.4 : 1};"
							role="button"
							tabindex="0"
							aria-label="{arc.label}: {arc.percent.toFixed(1)}%"
							onmouseenter={() => (hoveredSector = arc.key)}
							onmouseleave={() => (hoveredSector = null)}
						/>
					{/each}
				</g>

				<!-- Center text -->
				<text
					x={center}
					y={center - 6}
					text-anchor="middle"
					class="fill-foreground text-sm font-semibold"
				>
					{sectorData.length}
				</text>
				<text
					x={center}
					y={center + 10}
					text-anchor="middle"
					class="fill-muted-foreground text-[10px]"
				>
					sectors
				</text>
			</svg>

			<!-- Tooltip -->
			{#if hoveredSector}
				{@const sector = sectorData.find((s) => s.key === hoveredSector)}
				{#if sector}
					<div
						class="pointer-events-none absolute z-10 rounded-lg border border-border bg-popover px-3 py-2 text-xs shadow-lg"
						style="left: {tooltipX}px; top: {tooltipY}px; transform: translate(-50%, -100%);"
					>
						<p class="font-semibold" style="color: {sector.color}">{sector.label}</p>
						<p class="mt-1 tabular-nums">{fmtLarge(sector.value)}</p>
						<p class="text-muted-foreground tabular-nums">{sector.percent.toFixed(1)}%</p>
					</div>
				{/if}
			{/if}
		</div>

		<!-- Legend -->
		<div class="grid w-full grid-cols-2 gap-x-4 gap-y-2 text-xs lg:grid-cols-1">
			{#each sectorData.slice(0, 8) as sector (sector.key)}
				<button
					class="flex items-center gap-2 rounded px-1 py-0.5 text-left transition-colors hover:bg-muted/50"
					onmouseenter={() => (hoveredSector = sector.key)}
					onmouseleave={() => (hoveredSector = null)}
				>
					<div class="size-2.5 shrink-0 rounded-sm" style="background: {sector.color};"></div>
					<span class="truncate">{sector.label}</span>
					<span class="ml-auto shrink-0 font-medium text-muted-foreground tabular-nums">
						{sector.percent.toFixed(1)}%
					</span>
				</button>
			{/each}
			{#if sectorData.length > 8}
				<p class="col-span-2 text-muted-foreground lg:col-span-1">
					+{sectorData.length - 8} more sectors
				</p>
			{/if}
		</div>
	</div>
</div>
