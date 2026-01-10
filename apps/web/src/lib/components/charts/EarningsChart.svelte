<script lang="ts">
	import type { Fundamental } from '$lib/gen/ntx/v1/common_pb';

	interface Props {
		fundamentals: Fundamental[];
		class?: string;
	}

	let { fundamentals, class: className = '' }: Props = $props();

	// Chart dimensions
	const width = 800;
	const height = 300;
	const padding = { top: 30, right: 80, bottom: 50, left: 80 };
	const chartWidth = width - padding.left - padding.right;
	const chartHeight = height - padding.top - padding.bottom;

	// Transform and deduplicate data
	let chartData = $derived.by(() => {
		if (!fundamentals || fundamentals.length === 0) return [];

		// Deduplicate by fiscal year, keeping the latest entry
		const uniqueMap = new Map<string, (typeof fundamentals)[0]>();
		for (const f of fundamentals) {
			if (f.profitAmount !== undefined) {
				uniqueMap.set(f.fiscalYear, f);
			}
		}

		const sorted = Array.from(uniqueMap.values())
			.sort((a, b) => a.fiscalYear.localeCompare(b.fiscalYear))
			.slice(-5);

		return sorted.map((f, i) => {
			const prevProfit = i > 0 ? sorted[i - 1].profitAmount : undefined;
			const currentProfit = f.profitAmount ?? 0;
			const growth =
				prevProfit !== undefined && prevProfit !== 0
					? ((currentProfit - prevProfit) / Math.abs(prevProfit)) * 100
					: undefined;

			return {
				fiscalYear: f.fiscalYear,
				profit: currentProfit / 10_000_000, // Convert to Crores
				growth: growth ?? 0
			};
		});
	});

	// Calculate scales
	let maxProfit = $derived(Math.max(...chartData.map((d) => d.profit), 1));
	let maxGrowth = $derived(Math.max(...chartData.map((d) => Math.abs(d.growth)), 50));
	let minGrowth = $derived(Math.min(...chartData.map((d) => d.growth), -50));

	// Bar dimensions
	let barWidth = $derived(chartData.length > 0 ? chartWidth / chartData.length / 2 : 40);
	let barSpacing = $derived(chartData.length > 0 ? chartWidth / chartData.length : 60);

	// Scale functions
	function yScaleProfit(value: number): number {
		return chartHeight - (value / (maxProfit * 1.1)) * chartHeight;
	}

	function yScaleGrowth(value: number): number {
		const range = Math.max(maxGrowth, Math.abs(minGrowth)) * 1.2;
		return chartHeight / 2 - (value / range) * (chartHeight / 2);
	}

	function xPosition(index: number): number {
		return index * barSpacing + barSpacing / 2;
	}

	// Generate line path for growth
	let growthLinePath = $derived.by(() => {
		if (chartData.length === 0) return '';
		return chartData
			.map((d, i) => {
				const x = xPosition(i);
				const y = yScaleGrowth(d.growth);
				return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
			})
			.join(' ');
	});

	// Y-axis ticks for profit (left)
	let profitTicks = $derived.by(() => {
		const step = (maxProfit * 1.1) / 4;
		return [0, step, step * 2, step * 3, maxProfit * 1.1].map((v) => ({
			value: v,
			y: yScaleProfit(v),
			label: v >= 1 ? `${v.toFixed(0)}Cr` : `${(v * 100).toFixed(0)}L`
		}));
	});

	// Y-axis ticks for growth (right)
	let growthTicks = $derived.by(() => {
		const range = Math.max(maxGrowth, Math.abs(minGrowth)) * 1.2;
		return [-range, -range / 2, 0, range / 2, range].map((v) => ({
			value: v,
			y: yScaleGrowth(v),
			label: `${v >= 0 ? '+' : ''}${v.toFixed(0)}%`
		}));
	});

	function formatProfit(value: number): string {
		if (value >= 1) return `${value.toFixed(1)} Cr`;
		return `${(value * 100).toFixed(0)} L`;
	}
</script>

{#if chartData.length > 0}
	<div class={className}>
		<svg viewBox="0 0 {width} {height}" class="w-full" preserveAspectRatio="xMidYMid meet">
			<g transform="translate({padding.left}, {padding.top})">
				<!-- Grid lines -->
				{#each profitTicks as tick (tick.value)}
					<line
						x1="0"
						y1={tick.y}
						x2={chartWidth}
						y2={tick.y}
						stroke="currentColor"
						stroke-opacity="0.08"
						stroke-dasharray="4 4"
					/>
				{/each}

				<!-- Zero line for growth -->
				<line
					x1="0"
					y1={yScaleGrowth(0)}
					x2={chartWidth}
					y2={yScaleGrowth(0)}
					stroke="currentColor"
					stroke-opacity="0.15"
					stroke-width="1"
				/>

				<!-- Bars for Profit -->
				{#each chartData as d, i (d.fiscalYear)}
					{@const barHeight = chartHeight - yScaleProfit(d.profit)}
					{@const x = xPosition(i) - barWidth / 2}
					{@const y = yScaleProfit(d.profit)}

					<rect
						{x}
						{y}
						width={barWidth}
						height={barHeight}
						rx="6"
						fill="var(--chart-5)"
						opacity="0.9"
						class="transition-opacity hover:opacity-100"
					/>
				{/each}

				<!-- Growth line -->
				{#if growthLinePath}
					<path
						d={growthLinePath}
						fill="none"
						stroke="var(--chart-2)"
						stroke-width="2.5"
						stroke-linecap="round"
						stroke-linejoin="round"
					/>

					<!-- Growth dots -->
					{#each chartData as d, i (d.fiscalYear)}
						{@const x = xPosition(i)}
						{@const y = yScaleGrowth(d.growth)}
						<circle
							cx={x}
							cy={y}
							r="5"
							fill="var(--chart-2)"
							stroke="var(--background)"
							stroke-width="2.5"
							class="hover:r-6 transition-all"
						/>
					{/each}
				{/if}

				<!-- Left Y-axis labels (Profit) -->
				{#each profitTicks as tick (tick.value)}
					<text
						x="-12"
						y={tick.y}
						text-anchor="end"
						dominant-baseline="middle"
						class="fill-muted-foreground text-xs font-medium"
					>
						{tick.label}
					</text>
				{/each}

				<!-- Right Y-axis labels (Growth %) -->
				{#each growthTicks as tick (tick.value)}
					<text
						x={chartWidth + 12}
						y={tick.y}
						text-anchor="start"
						dominant-baseline="middle"
						class="fill-muted-foreground text-xs font-medium"
					>
						{tick.label}
					</text>
				{/each}

				<!-- X-axis labels -->
				{#each chartData as d, i (d.fiscalYear)}
					<text
						x={xPosition(i)}
						y={chartHeight + 25}
						text-anchor="middle"
						class="fill-foreground text-xs font-medium"
					>
						{d.fiscalYear}
					</text>
				{/each}
			</g>
		</svg>

		<!-- Legend -->
		<div class="mt-5 flex items-center justify-center gap-6 text-xs">
			<div class="flex items-center gap-2">
				<div class="h-3 w-3 rounded" style="background: var(--chart-5);"></div>
				<span class="text-muted-foreground">Net Profit</span>
			</div>
			<div class="flex items-center gap-2">
				<div class="h-0.5 w-4 rounded" style="background: var(--chart-2);"></div>
				<div class="h-2 w-2 rounded-full" style="background: var(--chart-2);"></div>
				<span class="text-muted-foreground">YoY Growth</span>
			</div>
		</div>
	</div>
{:else}
	<div class="flex h-[300px] items-center justify-center rounded-lg bg-muted/50">
		<p class="text-sm text-muted-foreground">No earnings data available</p>
	</div>
{/if}
