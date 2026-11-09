<script lang="ts">
	import { AreaChart } from 'layerchart';
	import { scaleTime, scaleLinear } from 'd3-scale';
	import type { Price } from '#lib/gen/ntx/v1/common_pb.js';
	import type { EventMark } from '#lib/gen/ntx/v1/story_pb.js';
	import { ChartContainer, type ChartConfig } from '#lib/components/ui/chart/index.js';

	interface Props {
		prices: Price[];
		events?: EventMark[];
		days?: number;
		class?: string;
	}

	let { prices, events = [], days = 365, class: className = '' }: Props = $props();

	// Filter data based on selected days and transform for chart
	const chartData = $derived.by(() => {
		if (!prices || prices.length === 0) return [];

		// Sort by date descending, then take the last N days
		const sorted = [...prices].sort((a, b) => b.businessDate.localeCompare(a.businessDate));

		const filtered = sorted.slice(0, days);

		// Transform to chart format and reverse for chronological order
		return filtered
			.map((p) => ({
				date: new Date(p.businessDate),
				price: p.ltp ?? p.close ?? 0,
				high: p.high ?? p.ltp ?? 0,
				low: p.low ?? p.ltp ?? 0,
				open: p.open ?? 0,
				close: p.close ?? 0,
				volume: Number(p.volume ?? 0)
			}))
			.reverse();
	});

	// Calculate price change color
	const priceDirection = $derived.by(() => {
		if (chartData.length < 2) return 'neutral';
		const first = chartData[0].price;
		const last = chartData[chartData.length - 1].price;
		if (last > first) return 'up';
		if (last < first) return 'down';
		return 'neutral';
	});

	// Calculate Y-axis domain with padding for better visibility
	const yDomain = $derived.by(() => {
		if (chartData.length === 0) return [0, 100];
		const prices = chartData.map((d) => d.price);
		const minPrice = Math.min(...prices);
		const maxPrice = Math.max(...prices);
		const padding = (maxPrice - minPrice) * 0.1 || maxPrice * 0.05;
		return [Math.max(0, minPrice - padding), maxPrice + padding];
	});

	// Notices label one row below results, so a result and a notice in the same week do not collide.
	const annotations = $derived.by(() => {
		if (chartData.length === 0) return [];
		const first = chartData[0].date.getTime();
		const last = chartData[chartData.length - 1].date.getTime();
		return events
			.map((e) => ({ e, date: new Date(e.date) }))
			.filter(({ date }) => date.getTime() >= first && date.getTime() <= last)
			.map(({ e, date }) => ({
				type: 'line' as const,
				x: date,
				label: e.label,
				labelPlacement: 'top' as const,
				labelYOffset: e.kind === 'result' ? 2 : -12,
				props: {
					line: {
						class: e.kind === 'result' ? 'stroke-chart-1/50' : 'stroke-chart-4/60',
						'stroke-dasharray': '3 3'
					},
					label: {
						class:
							e.kind === 'result' ? 'fill-muted-foreground text-[10px]' : 'fill-chart-4 text-[10px]'
					}
				}
			}));
	});

	const chartConfig: ChartConfig = {
		price: {
			label: 'Price',
			color: 'var(--chart-1)'
		}
	};
</script>

{#if chartData.length > 0}
	<ChartContainer config={chartConfig} class="w-full {className || 'h-[350px]'}">
		<AreaChart
			data={chartData}
			x="date"
			y="price"
			xScale={scaleTime()}
			yScale={scaleLinear().domain(yDomain).nice()}
			yDomain={yDomain as [number, number]}
			yBaseline={undefined}
			padding={{ top: 18, bottom: 30, left: 45, right: 15 }}
			{annotations}
			series={[
				{
					key: 'price',
					label: 'Price',
					value: (d: (typeof chartData)[number]) => d.price,
					color:
						priceDirection === 'up'
							? 'var(--positive)'
							: priceDirection === 'down'
								? 'var(--negative)'
								: 'var(--chart-1)'
				}
			]}
			props={{
				area: {
					line: {
						class:
							priceDirection === 'up'
								? 'stroke-positive stroke-2'
								: priceDirection === 'down'
									? 'stroke-negative stroke-2'
									: 'stroke-chart-1 stroke-2'
					},
					class:
						priceDirection === 'up'
							? 'fill-positive/15'
							: priceDirection === 'down'
								? 'fill-negative/15'
								: 'fill-chart-1/20'
				},
				grid: { class: 'stroke-border/20' }
			}}
		/>
	</ChartContainer>
{:else}
	<div class="flex h-[350px] items-center justify-center rounded-xl bg-muted/50">
		<div class="text-center">
			<svg
				class="mx-auto h-10 w-10 text-muted-foreground/50"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="1.5"
					d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z"
				/>
			</svg>
			<p class="mt-2 text-sm text-muted-foreground">No price data available</p>
		</div>
	</div>
{/if}
