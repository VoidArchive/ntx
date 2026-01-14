<script lang="ts">
	import type { Price } from '$lib/gen/ntx/v1/common_pb';

	interface Props {
		prices: Price[];
		days?: number;
		class?: string;
	}

	let { prices, days = 365, class: className = '' }: Props = $props();

	// Match PriceChart padding: { top: 10, bottom: 30, left: 45, right: 15 }
	const width = 800;
	const height = 80;
	const padding = { top: 5, bottom: 5, left: 45, right: 15 };
	const chartWidth = width - padding.left - padding.right;
	const chartHeight = height - padding.top - padding.bottom;

	// Filter and transform data
	const chartData = $derived.by(() => {
		if (!prices || prices.length === 0) return [];

		const sorted = [...prices].sort((a, b) => b.businessDate.localeCompare(a.businessDate));
		const filtered = sorted.slice(0, days);

		return filtered
			.map((p) => ({
				date: p.businessDate,
				volume: Number(p.volume ?? 0)
			}))
			.filter((d) => d.volume > 0)
			.reverse();
	});

	const maxVolume = $derived(Math.max(...chartData.map((d) => d.volume), 1));

	function yScale(value: number): number {
		return chartHeight - (value / (maxVolume * 1.1)) * chartHeight;
	}

	function barHeight(value: number): number {
		return Math.max(1, chartHeight - yScale(value));
	}

	function formatVolume(value: number): string {
		if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
		if (value >= 1_000) return `${(value / 1_000).toFixed(0)}K`;
		return value.toString();
	}

	// Y-axis ticks (just 2 labels: max and 0)
	const yTicks = $derived.by(() => {
		const max = maxVolume * 1.1;
		return [
			{ value: max, y: yScale(max), label: formatVolume(max) },
			{ value: 0, y: yScale(0), label: '0' }
		];
	});

	let hoveredIndex = $state<number | null>(null);
</script>

{#if chartData.length > 0}
	<div class="relative flex {className}">
		<!-- Y-axis labels (HTML, not stretched) -->
		<div class="flex h-[100px] w-[40px] shrink-0 flex-col justify-between pr-2 text-right text-[10px] text-muted-foreground tabular-nums">
			<span>{formatVolume(maxVolume)}</span>
			<span>0</span>
		</div>

		<!-- Chart bars -->
		<div class="relative flex-1">
			<svg viewBox="0 0 {chartWidth} {chartHeight}" class="h-[100px] w-full" preserveAspectRatio="none">
				{#each chartData as d, i (d.date)}
					{@const barW = chartWidth / chartData.length}
					{@const x = i * barW}
					{@const h = barHeight(d.volume)}
					{@const y = chartHeight - h}
					<rect
						{x}
						{y}
						width={Math.max(1, barW * 0.7)}
						height={h}
						class="fill-chart-2/60 transition-colors hover:fill-chart-2"
						role="button"
						tabindex={0}
						onmouseenter={() => (hoveredIndex = i)}
						onmouseleave={() => (hoveredIndex = null)}
					/>
				{/each}
			</svg>

			<!-- Tooltip -->
			{#if hoveredIndex !== null}
				{@const d = chartData[hoveredIndex]}
				{@const pctLeft = ((hoveredIndex + 0.5) / chartData.length) * 100}
				<div
					class="pointer-events-none absolute top-0 -translate-x-1/2 animate-in fade-in-0 zoom-in-95"
					style="left: {pctLeft}%;"
				>
					<div class="whitespace-nowrap rounded border border-border bg-popover px-2 py-1 text-xs shadow-md">
						<span class="font-medium tabular-nums">{formatVolume(d.volume)}</span>
					</div>
				</div>
			{/if}
		</div>
	</div>
{:else}
	<div class="flex h-[100px] items-center justify-center rounded-lg bg-muted/30">
		<p class="text-xs text-muted-foreground">No volume data</p>
	</div>
{/if}
