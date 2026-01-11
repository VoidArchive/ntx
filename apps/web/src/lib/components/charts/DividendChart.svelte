<script lang="ts">
	import type { CorporateAction } from '$lib/gen/ntx/v1/common_pb';
	import { SvelteMap } from 'svelte/reactivity';

	interface Props {
		actions: CorporateAction[];
		class?: string;
	}

	let { actions, class: className = '' }: Props = $props();

	const width = 700;
	const height = 280;
	const padding = { top: 30, right: 50, bottom: 50, left: 50 };
	const chartWidth = width - padding.left - padding.right;
	const chartHeight = height - padding.top - padding.bottom;

	let hoveredIndex = $state<number | null>(null);

	let chartData = $derived.by(() => {
		if (!actions || actions.length === 0) return [];

		const uniqueMap = new SvelteMap<string, (typeof actions)[0]>();
		for (const a of actions) {
			if (!uniqueMap.has(a.fiscalYear) || a.bonusPercentage || a.cashDividend) {
				uniqueMap.set(a.fiscalYear, a);
			}
		}

		return Array.from(uniqueMap.values())
			.sort((a, b) => a.fiscalYear.localeCompare(b.fiscalYear))
			.slice(-6)
			.map((a) => ({
				fiscalYear: a.fiscalYear,
				bonus: a.bonusPercentage ?? 0,
				cash: a.cashDividend ?? 0,
				total: (a.bonusPercentage ?? 0) + (a.cashDividend ?? 0)
			}));
	});

	let maxValue = $derived(
		Math.max(...chartData.map((d) => Math.max(d.bonus, d.cash, d.total)), 20)
	);
	let barWidth = $derived(
		chartData.length > 0 ? Math.min(chartWidth / chartData.length / 3, 36) : 36
	);
	let barSpacing = $derived(chartData.length > 0 ? chartWidth / chartData.length : 60);

	function yScale(value: number): number {
		return chartHeight - (value / (maxValue * 1.15)) * chartHeight;
	}

	function xPosition(index: number): number {
		return index * barSpacing + barSpacing / 2;
	}

	// S-curve bezier interpolation
	function generateSmoothPath(points: { x: number; y: number }[]): string {
		if (points.length < 2) return '';
		if (points.length === 2) {
			const midX = (points[0].x + points[1].x) / 2;
			return `M ${points[0].x} ${points[0].y} C ${midX} ${points[0].y}, ${midX} ${points[1].y}, ${points[1].x} ${points[1].y}`;
		}

		let path = `M ${points[0].x} ${points[0].y}`;

		for (let i = 0; i < points.length - 1; i++) {
			const p1 = points[i];
			const p2 = points[i + 1];

			const dx = p2.x - p1.x;
			const cp1x = p1.x + dx * 0.4;
			const cp2x = p1.x + dx * 0.6;

			path += ` C ${cp1x} ${p1.y}, ${cp2x} ${p2.y}, ${p2.x} ${p2.y}`;
		}

		return path;
	}

	let cashLinePath = $derived.by(() => {
		if (chartData.length === 0) return '';
		const points = chartData.map((d, i) => ({
			x: xPosition(i),
			y: yScale(d.cash)
		}));
		return generateSmoothPath(points);
	});

	let yTicks = $derived.by(() => {
		const max = maxValue * 1.15;
		const step = max / 4;
		return [0, step, step * 2, step * 3, max].map((v) => ({
			value: v,
			y: yScale(v),
			label: `${v.toFixed(0)}%`
		}));
	});

	function handleHover(index: number | null) {
		hoveredIndex = index;
	}
</script>

{#if chartData.length > 0}
	<div class="relative {className}">
		<svg viewBox="0 0 {width} {height}" class="w-full" preserveAspectRatio="xMidYMid meet">
			<defs>
				<linearGradient id="cashAreaGradient" x1="0" y1="0" x2="0" y2="1">
					<stop offset="0%" stop-color="#f97316" stop-opacity="0.2" />
					<stop offset="100%" stop-color="#f97316" stop-opacity="0" />
				</linearGradient>
			</defs>

			<g transform="translate({padding.left}, {padding.top})">
				<!-- Grid -->
				{#each yTicks as tick, i (i)}
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

				<!-- Cash area -->
				{#if cashLinePath}
					<path
						d="{cashLinePath} L {xPosition(chartData.length - 1)} {chartHeight} L {xPosition(
							0
						)} {chartHeight} Z"
						fill="url(#cashAreaGradient)"
					/>
				{/if}

				<!-- Bonus bars -->
				{#each chartData as d, i (d.fiscalYear)}
					{@const barHeight = Math.max(0, chartHeight - yScale(d.bonus))}
					{@const x = xPosition(i) - barWidth / 2}
					{@const y = yScale(d.bonus)}
					{@const isHovered = hoveredIndex === i}

					<rect
						{x}
						{y}
						width={barWidth}
						height={barHeight}
						rx="4"
						class="fill-orange-500 transition-opacity duration-150"
						opacity={hoveredIndex === null || isHovered ? 0.85 : 0.3}
					/>
				{/each}

				<!-- Cash line -->
				{#if cashLinePath}
					<path
						d={cashLinePath}
						fill="none"
						stroke="#fb923c"
						stroke-width="2.5"
						stroke-linecap="round"
						stroke-linejoin="round"
					/>

					{#each chartData as d, i (d.fiscalYear)}
						{@const x = xPosition(i)}
						{@const y = yScale(d.cash)}
						{@const isHovered = hoveredIndex === i}

						<circle
							cx={x}
							cy={y}
							r={isHovered ? 6 : 4}
							class="fill-orange-400 transition-all duration-150"
							stroke="var(--background)"
							stroke-width="2"
						/>
					{/each}
				{/if}

				<!-- Hover areas -->
				{#each chartData as d, i (d.fiscalYear)}
					<rect
						x={xPosition(i) - barSpacing / 2}
						y="0"
						width={barSpacing}
						height={chartHeight}
						fill="transparent"
						role="button"
						tabindex="0"
						aria-label="{d.fiscalYear}: Bonus {d.bonus}%, Cash {d.cash}%"
						class="cursor-pointer"
						onmouseenter={() => handleHover(i)}
						onmouseleave={() => handleHover(null)}
						onfocus={() => handleHover(i)}
						onblur={() => handleHover(null)}
					/>
				{/each}

				<!-- X-axis -->
				{#each chartData as d, i (d.fiscalYear)}
					<text
						x={xPosition(i)}
						y={chartHeight + 20}
						text-anchor="middle"
						class="fill-muted-foreground text-[10px]"
					>
						{d.fiscalYear}
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
					<p class="text-xs font-medium">{d.fiscalYear}</p>
					<div class="mt-1 flex gap-3 text-[11px]">
						<span class="text-orange-500">Bonus: {d.bonus}%</span>
						<span class="text-orange-400">Cash: {d.cash}%</span>
					</div>
					{#if d.total > 0}
						<p class="mt-1 text-[10px] text-muted-foreground">Total: {d.total}%</p>
					{/if}
				</div>
			</div>
		{/if}

		<!-- Legend -->
		<div class="mt-4 flex items-center justify-center gap-6 text-xs">
			<div class="flex items-center gap-2">
				<div class="h-3 w-3 rounded bg-orange-500"></div>
				<span class="text-muted-foreground">Bonus</span>
			</div>
			<div class="flex items-center gap-2">
				<div class="h-0.5 w-4 rounded-full bg-orange-400"></div>
				<div class="h-2 w-2 rounded-full bg-orange-400"></div>
				<span class="text-muted-foreground">Cash</span>
			</div>
		</div>
	</div>
{:else}
	<div
		class="flex h-[280px] items-center justify-center rounded-xl border border-dashed border-border"
	>
		<p class="text-sm text-muted-foreground">No dividend data available</p>
	</div>
{/if}
