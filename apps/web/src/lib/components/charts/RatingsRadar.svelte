<script lang="ts">
	import type { Fundamental, Price } from '$lib/gen/ntx/v1/common_pb';
	import { SvelteMap } from 'svelte/reactivity';

	interface Props {
		fundamentals: Fundamental[];
		price?: Price;
		class?: string;
	}

	let { fundamentals, price, class: className = '' }: Props = $props();

	function quarterToNumber(quarter: string | undefined): number {
		if (!quarter) return 5;
		const q = quarter.toLowerCase();
		if (q.includes('fourth') || q === '4' || q === 'q4') return 4;
		if (q.includes('third') || q === '3' || q === 'q3') return 3;
		if (q.includes('second') || q === '2' || q === 'q2') return 2;
		if (q.includes('first') || q === '1' || q === 'q1') return 1;
		return 0;
	}

	// Deduplicate and sort fundamentals by fiscal year and quarter descending
	let sortedFundamentals = $derived.by(() => {
		const seen = new SvelteMap<string, (typeof fundamentals)[0]>();
		for (const f of fundamentals) {
			const key = `${f.fiscalYear}-${f.quarter ?? 'annual'}`;
			if (!seen.has(key)) {
				seen.set(key, f);
			}
		}
		return Array.from(seen.values()).sort((a, b) => {
			const yearCompare = b.fiscalYear.localeCompare(a.fiscalYear);
			if (yearCompare !== 0) return yearCompare;
			return quarterToNumber(b.quarter) - quarterToNumber(a.quarter);
		});
	});

	// Get latest fundamental from passed data (already filtered by page)
	let latestFundamental = $derived(sortedFundamentals[0]);

	// Get previous entry for growth calculations (YoY for quarterly data)
	let previousFundamental = $derived.by(() => {
		if (!latestFundamental) return undefined;

		// For quarterly data, find same quarter from previous year
		if (latestFundamental.quarter) {
			return sortedFundamentals.find(
				(f) =>
					f.quarter === latestFundamental.quarter && f.fiscalYear !== latestFundamental.fiscalYear
			);
		}

		// For annual data, just use the previous year
		return sortedFundamentals[1];
	});

	// Calculate metrics and ratings (1-5 scale)
	let metrics = $derived.by(() => {
		if (!latestFundamental) return null;

		const ltp = price?.ltp ?? price?.close ?? 0;
		const eps = latestFundamental.eps ?? 0;
		const bookValue = latestFundamental.bookValue ?? 0;
		const pe = latestFundamental.peRatio ?? (eps > 0 ? ltp / eps : 0);
		const pb = bookValue > 0 ? ltp / bookValue : 0;
		const roe = bookValue > 0 ? (eps / bookValue) * 100 : 0;

		// Growth calculations
		const prevEps = previousFundamental?.eps ?? 0;
		const prevProfit = previousFundamental?.profitAmount ?? 0;
		const currentProfit = latestFundamental.profitAmount ?? 0;

		const epsGrowth = prevEps > 0 ? ((eps - prevEps) / prevEps) * 100 : 0;
		const profitGrowth = prevProfit > 0 ? ((currentProfit - prevProfit) / prevProfit) * 100 : 0;

		// Rating functions (return 1-5)
		function ratePE(v: number): number {
			if (v <= 0) return 0;
			if (v < 10) return 5;
			if (v < 15) return 4;
			if (v < 25) return 3;
			if (v < 40) return 2;
			return 1;
		}

		function ratePB(v: number): number {
			if (v <= 0) return 0;
			if (v < 1) return 5;
			if (v < 2) return 4;
			if (v < 3) return 3;
			if (v < 5) return 2;
			return 1;
		}

		function rateROE(v: number): number {
			if (v <= 0) return 1;
			if (v > 20) return 5;
			if (v > 15) return 4;
			if (v > 10) return 3;
			if (v > 5) return 2;
			return 1;
		}

		function rateGrowth(v: number): number {
			if (v > 30) return 5;
			if (v > 15) return 4;
			if (v > 0) return 3;
			if (v > -15) return 2;
			return 1;
		}

		const peRating = ratePE(pe);
		const pbRating = ratePB(pb);
		const roeRating = rateROE(roe);
		const epsGrowthRating = rateGrowth(epsGrowth);
		const profitGrowthRating = rateGrowth(profitGrowth);

		const validRatings = [
			peRating,
			pbRating,
			roeRating,
			epsGrowthRating,
			profitGrowthRating
		].filter((r) => r > 0);
		const overall =
			validRatings.length > 0 ? validRatings.reduce((a, b) => a + b, 0) / validRatings.length : 0;

		return {
			pe: { value: pe, rating: peRating, label: 'P/E' },
			pb: { value: pb, rating: pbRating, label: 'P/B' },
			roe: { value: roe, rating: roeRating, label: 'ROE' },
			epsGrowth: { value: epsGrowth, rating: epsGrowthRating, label: 'EPS Gr.' },
			profitGrowth: { value: profitGrowth, rating: profitGrowthRating, label: 'Profit Gr.' },
			overall
		};
	});

	// Radar chart geometry
	const size = 200;
	const center = size / 2;
	const radius = 70;
	const levels = 5;

	// 5 axes for pentagon
	const axes = ['pe', 'roe', 'profitGrowth', 'pb', 'epsGrowth'] as const;

	function getPoint(index: number, value: number): { x: number; y: number } {
		const angle = (Math.PI * 2 * index) / axes.length - Math.PI / 2;
		const r = (value / 5) * radius;
		return {
			x: center + r * Math.cos(angle),
			y: center + r * Math.sin(angle)
		};
	}

	function getLabelPoint(index: number): { x: number; y: number } {
		const angle = (Math.PI * 2 * index) / axes.length - Math.PI / 2;
		const r = radius + 24;
		return {
			x: center + r * Math.cos(angle),
			y: center + r * Math.sin(angle)
		};
	}

	let radarPath = $derived.by(() => {
		if (!metrics) return '';
		return (
			axes
				.map((axis, i) => {
					const point = getPoint(i, metrics[axis].rating);
					return `${i === 0 ? 'M' : 'L'} ${point.x} ${point.y}`;
				})
				.join(' ') + ' Z'
		);
	});

	function getGrade(score: number): string {
		if (score >= 4.5) return 'A+';
		if (score >= 4) return 'A';
		if (score >= 3.5) return 'B+';
		if (score >= 3) return 'B';
		if (score >= 2.5) return 'C+';
		if (score >= 2) return 'C';
		return 'D';
	}

	function formatValue(axis: string, value: number): string {
		if (axis === 'pe' || axis === 'pb') return value.toFixed(1);
		return `${value.toFixed(0)}%`;
	}
</script>

{#if metrics}
	<div class="flex flex-col items-center gap-4 {className}">
		<!-- Header -->
		<div class="text-center">
			<h3 class="font-serif text-base font-medium">Ratings Snapshot</h3>
			<p class="mt-1 text-sm">
				<span class="text-muted-foreground">Rating:</span>
				<span class="ml-1 font-semibold text-orange-500">{getGrade(metrics.overall)}</span>
			</p>
		</div>

		<div class="flex flex-col items-center gap-6 sm:flex-row sm:items-start">
			<!-- Radar Chart -->
			<svg viewBox="0 0 {size} {size}" class="h-40 w-40 sm:h-48 sm:w-48">
				<!-- Grid levels -->
				{#each Array(levels) as _, level (level)}
					{@const r = ((level + 1) / levels) * radius}
					<polygon
						points={axes
							.map((_a, i) => {
								const angle = (Math.PI * 2 * i) / axes.length - Math.PI / 2;
								return `${center + r * Math.cos(angle)},${center + r * Math.sin(angle)}`;
							})
							.join(' ')}
						fill="none"
						stroke="currentColor"
						stroke-opacity="0.1"
						stroke-width="1"
					/>
				{/each}

				<!-- Axis lines -->
				{#each axes as _axis, i (i)}
					{@const point = getPoint(i, 5)}
					<line
						x1={center}
						y1={center}
						x2={point.x}
						y2={point.y}
						stroke="currentColor"
						stroke-opacity="0.1"
						stroke-width="1"
					/>
				{/each}

				<!-- Data polygon -->
				<path
					d={radarPath}
					fill="#f97316"
					fill-opacity="0.2"
					stroke="#f97316"
					stroke-width="2"
					stroke-linejoin="round"
				/>

				<!-- Data points -->
				{#each axes as axis, i (axis)}
					{@const point = getPoint(i, metrics[axis].rating)}
					<circle cx={point.x} cy={point.y} r="4" fill="#f97316" />
				{/each}

				<!-- Labels -->
				{#each axes as axis, i (axis)}
					{@const labelPoint = getLabelPoint(i)}
					<text
						x={labelPoint.x}
						y={labelPoint.y}
						text-anchor="middle"
						dominant-baseline="middle"
						class="fill-muted-foreground text-[10px] font-medium"
					>
						{metrics[axis].label}
					</text>
				{/each}
			</svg>

			<!-- Values list -->
			<div class="space-y-2 text-sm">
				{#each axes as axis (axis)}
					{@const m = metrics[axis]}
					<div class="flex items-center gap-3">
						<span class="w-16 text-muted-foreground">{m.label}</span>
						<span class="w-12 text-right font-medium tabular-nums"
							>{formatValue(axis, m.value)}</span
						>
						<span class="w-4 text-center font-semibold text-orange-500">{m.rating}</span>
					</div>
				{/each}
				<div class="border-t border-border pt-2">
					<div class="flex items-center gap-3 font-medium">
						<span class="w-16">Overall</span>
						<span class="w-12 text-right text-orange-500 tabular-nums"
							>{metrics.overall.toFixed(1)}</span
						>
					</div>
				</div>
			</div>
		</div>
	</div>
{:else}
	<div class="flex h-48 items-center justify-center rounded-xl border border-dashed border-border">
		<p class="text-sm text-muted-foreground">Insufficient data for ratings</p>
	</div>
{/if}
