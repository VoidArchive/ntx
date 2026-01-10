<script lang="ts">
	import { AreaChart } from 'layerchart';
	import { scaleTime, scaleLinear } from 'd3-scale';
	import type { Price } from '$lib/gen/ntx/v1/common_pb';
	import { ChartContainer, type ChartConfig } from '$lib/components/ui/chart';

	interface Props {
		prices: Price[];
		days?: 30 | 90 | 180 | 365;
		class?: string;
	}

	let { prices, days = 365, class: className = '' }: Props = $props();

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

	const chartConfig: ChartConfig = {
		price: {
			label: 'Price',
			color: 'var(--chart-1)'
		}
	};
</script>

{#if chartData.length > 0}
	<ChartContainer config={chartConfig} class="h-[350px] w-full {className}">
		<AreaChart
			data={chartData}
			x="date"
			y="price"
			xScale={scaleTime()}
			yScale={scaleLinear().nice()}
			series={[
				{
					key: 'price',
					value: (d) => d.price,
					color:
						priceDirection === 'up'
							? 'var(--positive)'
							: priceDirection === 'down'
								? 'var(--negative)'
								: 'var(--chart-1)'
				}
			]}
			tooltip={{ title: 'Price' }}
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
