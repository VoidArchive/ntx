<script lang="ts">
	import type { Fundamental } from '$lib/gen/ntx/v1/common_pb';
	import { SvelteMap } from 'svelte/reactivity';

	interface Props {
		fundamentals: Fundamental[];
		class?: string;
	}

	let { fundamentals, class: className = '' }: Props = $props();

	const width = 700;
	const height = 280;
	const padding = { top: 30, right: 60, bottom: 50, left: 60 };
	const chartWidth = width - padding.left - padding.right;
	const chartHeight = height - padding.top - padding.bottom;

	let hoveredIndex = $state<number | null>(null);

	// Convert "First Quarter" → "Q1", "Second Quarter" → "Q2", etc.
	function abbreviateQuarter(quarter: string): string {
		const map: Record<string, string> = {
			'First Quarter': 'Q1',
			'Second Quarter': 'Q2',
			'Third Quarter': 'Q3',
			'Fourth Quarter': 'Q4'
		};
		return map[quarter] ?? quarter;
	}

	// Shorten fiscal year: "2023-2024" → "23/24", "2081" → "81"
	function shortenFiscalYear(year: string): string {
		if (year.includes('-')) {
			const [start, end] = year.split('-');
			return `${start.slice(-2)}/${end.slice(-2)}`;
		}
		return year.slice(-2);
	}

	let chartData = $derived.by(() => {
		if (!fundamentals || fundamentals.length === 0) return [];

		const uniqueMap = new SvelteMap<string, (typeof fundamentals)[0]>();
		for (const f of fundamentals) {
			if (f.profitAmount !== undefined) {
				// Use fiscalYear + quarter as key for quarterly data
				const key = f.quarter ? `${f.fiscalYear}-${f.quarter}` : f.fiscalYear;
				uniqueMap.set(key, f);
			}
		}

		const sorted = Array.from(uniqueMap.values())
			.sort((a, b) => {
				const yearCmp = a.fiscalYear.localeCompare(b.fiscalYear);
				if (yearCmp !== 0) return yearCmp;
				// Sort quarters: Q1 < Q2 < Q3 < Q4
				return (a.quarter ?? '').localeCompare(b.quarter ?? '');
			})
			.slice(-6);

		return sorted.map((f, i) => {
			const prevProfit = i > 0 ? sorted[i - 1].profitAmount : undefined;
			const currentProfit = f.profitAmount ?? 0;

			// Only calculate growth if previous profit is meaningful (> 1% of current)
			// This avoids extreme percentages when previous was near zero
			let growth: number | null = null;
			if (prevProfit !== undefined && prevProfit !== 0) {
				const rawGrowth = ((currentProfit - prevProfit) / Math.abs(prevProfit)) * 100;
				// Cap at ±500% to keep chart readable
				if (Math.abs(rawGrowth) <= 500) {
					growth = rawGrowth;
				}
			}

			// Build label: "Q1 23/24" for quarterly, "23/24" for annual
			const shortYear = shortenFiscalYear(f.fiscalYear);
			const label = f.quarter ? `${abbreviateQuarter(f.quarter)} ${shortYear}` : shortYear;
			// Full label for tooltip: "Q1 2023-2024" or "2023-2024"
			const fullLabel = f.quarter
				? `${abbreviateQuarter(f.quarter)} ${f.fiscalYear}`
				: f.fiscalYear;

			return {
				fiscalYear: f.fiscalYear,
				quarter: f.quarter,
				label,
				fullLabel,
				profit: currentProfit / 10_000_000,
				growth
			};
		});
	});

	let maxProfit = $derived(Math.max(...chartData.map((d) => d.profit), 1));
	let maxGrowth = $derived(Math.max(...chartData.map((d) => Math.abs(d.growth ?? 0)), 30));

	let barWidth = $derived(
		chartData.length > 0 ? Math.min(chartWidth / chartData.length / 3, 36) : 36
	);
	let barSpacing = $derived(chartData.length > 0 ? chartWidth / chartData.length : 60);

	function yScaleProfit(value: number): number {
		return chartHeight - (value / (maxProfit * 1.15)) * chartHeight;
	}

	function yScaleGrowth(value: number): number {
		const range = maxGrowth * 1.3;
		return chartHeight / 2 - (value / range) * (chartHeight / 2);
	}

	function xPosition(index: number): number {
		return index * barSpacing + barSpacing / 2;
	}

	// Attempt monotone cubic interpolation for smooth S-curves
	function generateSmoothPath(points: { x: number; y: number }[]): string {
		if (points.length < 2) return '';
		if (points.length === 2) {
			// Simple S-curve between two points
			const midX = (points[0].x + points[1].x) / 2;
			return `M ${points[0].x} ${points[0].y} C ${midX} ${points[0].y}, ${midX} ${points[1].y}, ${points[1].x} ${points[1].y}`;
		}

		let path = `M ${points[0].x} ${points[0].y}`;

		for (let i = 0; i < points.length - 1; i++) {
			const p1 = points[i];
			const p2 = points[i + 1];

			// Control points at 40% and 60% of the horizontal distance
			// but keeping Y values from start/end for S-curve effect
			const dx = p2.x - p1.x;
			const cp1x = p1.x + dx * 0.4;
			const cp2x = p1.x + dx * 0.6;

			path += ` C ${cp1x} ${p1.y}, ${cp2x} ${p2.y}, ${p2.x} ${p2.y}`;
		}

		return path;
	}

	let growthLinePath = $derived.by(() => {
		const validPoints = chartData
			.map((d, i) => (d.growth !== null ? { x: xPosition(i), y: yScaleGrowth(d.growth) } : null))
			.filter((p): p is { x: number; y: number } => p !== null);
		return generateSmoothPath(validPoints);
	});

	let profitTicks = $derived.by(() => {
		const max = maxProfit * 1.15;
		const step = max / 4;
		return [0, step, step * 2, step * 3, max].map((v) => ({
			value: v,
			y: yScaleProfit(v),
			label: v >= 1 ? `${v.toFixed(0)}Cr` : `${(v * 100).toFixed(0)}L`
		}));
	});

	let growthTicks = $derived.by(() => {
		const range = maxGrowth * 1.3;
		return [-range, 0, range].map((v) => ({
			value: v,
			y: yScaleGrowth(v),
			label: `${v >= 0 ? '+' : ''}${v.toFixed(0)}%`
		}));
	});

	function formatProfit(value: number): string {
		if (Math.abs(value) >= 1) return `${value.toFixed(1)} Cr`;
		return `${(value * 100).toFixed(0)} L`;
	}

	function handleHover(index: number | null) {
		hoveredIndex = index;
	}
</script>

{#if chartData.length > 0}
	<div class="relative {className}">
		<svg viewBox="0 0 {width} {height}" class="w-full" preserveAspectRatio="xMidYMid meet">
			<defs>
				<linearGradient id="growthAreaGradient" x1="0" y1="0" x2="0" y2="1">
					<stop offset="0%" stop-color="#0ea5e9" stop-opacity="0.15" />
					<stop offset="100%" stop-color="#0ea5e9" stop-opacity="0" />
				</linearGradient>
			</defs>

			<g transform="translate({padding.left}, {padding.top})">
				<!-- Grid -->
				{#each profitTicks as tick, i (i)}
					<line
						x1="0"
						y1={tick.y}
						x2={chartWidth}
						y2={tick.y}
						stroke="currentColor"
						stroke-opacity="0.06"
					/>
					<text
						x="-8"
						y={tick.y}
						text-anchor="end"
						dominant-baseline="middle"
						class="fill-muted-foreground text-[10px] tabular-nums"
					>
						{tick.label}
					</text>
				{/each}

				<!-- Growth zero line -->
				<line
					x1="0"
					y1={yScaleGrowth(0)}
					x2={chartWidth}
					y2={yScaleGrowth(0)}
					stroke="#0ea5e9"
					stroke-opacity="0.2"
					stroke-dasharray="4 4"
				/>

				<!-- Right axis labels -->
				{#each growthTicks as tick (tick.value)}
					<text
						x={chartWidth + 8}
						y={tick.y}
						text-anchor="start"
						dominant-baseline="middle"
						class="fill-muted-foreground text-[10px] tabular-nums"
					>
						{tick.label}
					</text>
				{/each}

				<!-- Growth area -->
				{#if growthLinePath}
					{@const validIndices = chartData
						.map((d, i) => (d.growth !== null ? i : -1))
						.filter((i) => i >= 0)}
					{@const firstIdx = validIndices[0]}
					{@const lastIdx = validIndices[validIndices.length - 1]}
					<path
						d="{growthLinePath} L {xPosition(lastIdx)} {yScaleGrowth(0)} L {xPosition(
							firstIdx
						)} {yScaleGrowth(0)} Z"
						fill="url(#growthAreaGradient)"
					/>
				{/if}

				<!-- Profit bars -->
				{#each chartData as d, i (d.label)}
					{@const barHeight = Math.max(0, chartHeight - yScaleProfit(d.profit))}
					{@const x = xPosition(i) - barWidth / 2}
					{@const y = yScaleProfit(d.profit)}
					{@const isHovered = hoveredIndex === i}

					<rect
						{x}
						{y}
						width={barWidth}
						height={barHeight}
						rx="4"
						class="fill-emerald-500 transition-opacity duration-150"
						opacity={hoveredIndex === null || isHovered ? 0.85 : 0.3}
					/>
				{/each}

				<!-- Growth line -->
				{#if growthLinePath}
					<path
						d={growthLinePath}
						fill="none"
						stroke="#0ea5e9"
						stroke-width="2.5"
						stroke-linecap="round"
						stroke-linejoin="round"
					/>

					{#each chartData as d, i (d.label)}
						{#if d.growth !== null}
							{@const x = xPosition(i)}
							{@const y = yScaleGrowth(d.growth)}
							{@const isHovered = hoveredIndex === i}

							<circle
								cx={x}
								cy={y}
								r={isHovered ? 6 : 4}
								class="fill-sky-500 transition-all duration-150"
								stroke="var(--background)"
								stroke-width="2"
							/>
						{/if}
					{/each}
				{/if}

				<!-- Hover areas -->
				{#each chartData as d, i (d.label)}
					<rect
						x={xPosition(i) - barSpacing / 2}
						y="0"
						width={barSpacing}
						height={chartHeight}
						fill="transparent"
						role="button"
						tabindex="0"
						aria-label="{d.label}: Profit {formatProfit(d.profit)}"
						class="cursor-pointer"
						onmouseenter={() => handleHover(i)}
						onmouseleave={() => handleHover(null)}
						onfocus={() => handleHover(i)}
						onblur={() => handleHover(null)}
					/>
				{/each}

				<!-- X-axis -->
				{#each chartData as d, i (d.label)}
					<text
						x={xPosition(i)}
						y={chartHeight + 20}
						text-anchor="middle"
						class="fill-muted-foreground text-[10px]"
					>
						{d.label}
					</text>
				{/each}
			</g>
		</svg>

		<!-- Tooltip -->
		{#if hoveredIndex !== null}
			{@const d = chartData[hoveredIndex]}
			{@const x = ((xPosition(hoveredIndex) + padding.left) / width) * 100}
			<div
				class="pointer-events-none absolute -translate-x-1/2 animate-in duration-100 fade-in-0 zoom-in-95"
				style="left: {x}%; top: 0;"
			>
				<div class="rounded-lg border border-border/50 bg-popover px-3 py-2 shadow-lg">
					<p class="text-xs font-medium">{d.fullLabel}</p>
					<div class="mt-1 flex gap-3 text-[11px]">
						<span class="text-emerald-500">Profit: {formatProfit(d.profit)}</span>
						{#if d.growth !== null}
							<span class={d.growth >= 0 ? 'text-sky-500' : 'text-red-500'}>
								{d.growth >= 0 ? '+' : ''}{d.growth.toFixed(1)}%
							</span>
						{/if}
					</div>
				</div>
			</div>
		{/if}

		<!-- Legend -->
		<div class="mt-4 flex items-center justify-center gap-6 text-xs">
			<div class="flex items-center gap-2">
				<div class="h-3 w-3 rounded bg-emerald-500"></div>
				<span class="text-muted-foreground">Net Profit</span>
			</div>
			<div class="flex items-center gap-2">
				<div class="h-0.5 w-4 rounded-full bg-sky-500"></div>
				<div class="h-2 w-2 rounded-full bg-sky-500"></div>
				<span class="text-muted-foreground">YoY Growth</span>
			</div>
		</div>
	</div>
{:else}
	<div
		class="flex h-[280px] items-center justify-center rounded-xl border border-dashed border-border"
	>
		<p class="text-sm text-muted-foreground">No earnings data available</p>
	</div>
{/if}
