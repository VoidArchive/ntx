<script lang="ts">
	import type { Ownership } from '$lib/gen/ntx/v1/common_pb';

	interface Props {
		ownership?: Ownership;
		class?: string;
	}

	let { ownership, class: className = '' }: Props = $props();

	// Chart dimensions
	const size = 180;
	const center = size / 2;
	const radius = 65;
	const innerRadius = 35; // Donut hole

	// Tailwind Orange palette
	const colors = {
		promoter: '#f97316', // orange-500
		promoterHover: '#ea580c', // orange-600
		public: '#fdba74', // orange-300
		publicHover: '#fb923c' // orange-400
	};

	// Calculate percentages
	let promoterPercent = $derived(ownership?.promoterPercent ?? 0);
	let publicPercent = $derived(ownership?.publicPercent ?? 0);
	let total = $derived(promoterPercent + publicPercent);
	let normalizedPromoter = $derived(total > 0 ? (promoterPercent / total) * 100 : 50);

	// Hover state
	let hoveredSlice = $state<'promoter' | 'public' | null>(null);

	// Create donut arc path
	function createArc(startPercent: number, endPercent: number): string {
		const startAngle = (startPercent / 100) * 360 - 90;
		const endAngle = (endPercent / 100) * 360 - 90;

		const outerStart = polarToCartesian(center, center, radius, endAngle);
		const outerEnd = polarToCartesian(center, center, radius, startAngle);
		const innerStart = polarToCartesian(center, center, innerRadius, endAngle);
		const innerEnd = polarToCartesian(center, center, innerRadius, startAngle);

		const largeArc = endPercent - startPercent > 50 ? 1 : 0;

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

	function polarToCartesian(cx: number, cy: number, r: number, angleInDegrees: number) {
		const rad = (angleInDegrees * Math.PI) / 180;
		return { x: cx + r * Math.cos(rad), y: cy + r * Math.sin(rad) };
	}

	// Arc paths
	let promoterArc = $derived(createArc(0, normalizedPromoter));
	let publicArc = $derived(createArc(normalizedPromoter, 100));

	function formatShares(shares: number | bigint): string {
		const num = Number(shares);
		if (num >= 10_000_000) return `${(num / 10_000_000).toFixed(2)} Cr`;
		if (num >= 100_000) return `${(num / 100_000).toFixed(2)} L`;
		return num.toLocaleString();
	}

	function formatSharesFull(shares: number | bigint): string {
		return Number(shares).toLocaleString();
	}

	// Tooltip position
	let tooltipX = $state(0);
	let tooltipY = $state(0);

	function handleMouseMove(e: MouseEvent) {
		const rect = (e.currentTarget as SVGElement).getBoundingClientRect();
		tooltipX = e.clientX - rect.left;
		tooltipY = e.clientY - rect.top - 10;
	}
</script>

{#if ownership && total > 0}
	<div class={className}>
		<h3 class="mb-4 font-serif text-base font-medium">Shareholding</h3>

		<div class="flex flex-col items-center">
			<!-- Donut Chart -->
			<div class="relative">
				<svg
					width={size}
					height={size}
					viewBox="0 0 {size} {size}"
					class="drop-shadow-md"
					role="img"
					aria-label="Shareholding pie chart"
					onmousemove={handleMouseMove}
				>
					<defs>
						<filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
							<feDropShadow dx="0" dy="2" stdDeviation="2" flood-opacity="0.15" />
						</filter>
					</defs>

					<g filter="url(#glow)">
						<!-- Promoter slice -->
						<path
							d={promoterArc}
							fill={hoveredSlice === 'promoter' ? colors.promoterHover : colors.promoter}
							class="cursor-pointer transition-all duration-150"
							role="button"
							tabindex="0"
							aria-label="Promoter shares: {promoterPercent.toFixed(1)}%"
							onmouseenter={() => (hoveredSlice = 'promoter')}
							onmouseleave={() => (hoveredSlice = null)}
						/>

						<!-- Public slice -->
						<path
							d={publicArc}
							fill={hoveredSlice === 'public' ? colors.publicHover : colors.public}
							class="cursor-pointer transition-all duration-150"
							role="button"
							tabindex="0"
							aria-label="Public shares: {publicPercent.toFixed(1)}%"
							onmouseenter={() => (hoveredSlice = 'public')}
							onmouseleave={() => (hoveredSlice = null)}
						/>
					</g>

					<!-- Center text -->
					<text
						x={center}
						y={center - 4}
						text-anchor="middle"
						class="fill-foreground text-sm font-semibold"
					>
						{formatShares(ownership.listedShares)}
					</text>
					<text
						x={center}
						y={center + 12}
						text-anchor="middle"
						class="fill-muted-foreground text-[10px]"
					>
						listed
					</text>
				</svg>

				<!-- Tooltip -->
				{#if hoveredSlice}
					<div
						class="pointer-events-none absolute z-10 rounded-lg border border-border bg-popover px-3 py-2 text-xs shadow-lg"
						style="left: {tooltipX}px; top: {tooltipY}px; transform: translate(-50%, -100%);"
					>
						{#if hoveredSlice === 'promoter'}
							<p class="font-semibold" style="color: {colors.promoter}">Promoter</p>
							<p class="mt-1 tabular-nums">{formatSharesFull(ownership.promoterShares)} shares</p>
							<p class="text-muted-foreground tabular-nums">{promoterPercent.toFixed(2)}%</p>
						{:else}
							<p class="font-semibold" style="color: {colors.promoter}">Public</p>
							<p class="mt-1 tabular-nums">{formatSharesFull(ownership.publicShares)} shares</p>
							<p class="text-muted-foreground tabular-nums">{publicPercent.toFixed(2)}%</p>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Legend -->
			<div class="mt-4 flex gap-6 text-xs">
				<div class="flex items-center gap-2">
					<div class="h-3 w-3 rounded-sm" style="background: {colors.promoter};"></div>
					<span>Promoter</span>
					<span class="font-semibold tabular-nums">{promoterPercent.toFixed(1)}%</span>
				</div>
				<div class="flex items-center gap-2">
					<div class="h-3 w-3 rounded-sm" style="background: {colors.public};"></div>
					<span>Public</span>
					<span class="font-semibold tabular-nums">{publicPercent.toFixed(1)}%</span>
				</div>
			</div>
		</div>
	</div>
{:else}
	<div class={className}>
		<h3 class="mb-4 font-serif text-base font-medium">Shareholding</h3>
		<p class="text-sm text-muted-foreground">No ownership data available</p>
	</div>
{/if}
