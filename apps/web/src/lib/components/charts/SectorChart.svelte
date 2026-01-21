<script lang="ts">
	interface SectorData {
		label: string;
		value: number; // percentage or value
		color: string;
	}

	interface Props {
		data: SectorData[];
		class?: string;
	}

	let { data, class: className = '' }: Props = $props();

	// Chart dimensions
	const size = 200;
	const center = size / 2;
	const radius = 70;
	const innerRadius = 40;

	// Calculate total (should be 100 if percentages, but let's be safe)
	let total = $derived(data.reduce((acc, d) => acc + d.value, 0));

	// Calculate paths
	let slices = $derived(() => {
		let currentAngle = -90; // Start from top
		return data.map((d) => {
			const percent = total > 0 ? d.value / total : 0;
			const angle = percent * 360;
			const startAngle = currentAngle;
			const endAngle = currentAngle + angle;
			currentAngle = endAngle;

			return {
				...d,
				percent: percent * 100,
				path: createArc(startAngle, endAngle)
			};
		});
	});

	function createArc(startAngle: number, endAngle: number): string {
		// Ensure angles are properly handled (mod 360 if needed, but simple add is fine here)
		// Convert to radians
		const startRad = (startAngle * Math.PI) / 180;
		const endRad = (endAngle * Math.PI) / 180;

		const largeArc = endAngle - startAngle > 180 ? 1 : 0;

		const outerStart = { x: center + radius * Math.cos(startRad), y: center + radius * Math.sin(startRad) };
		const outerEnd = { x: center + radius * Math.cos(endRad), y: center + radius * Math.sin(endRad) };
		const innerStart = { x: center + innerRadius * Math.cos(startRad), y: center + innerRadius * Math.sin(startRad) };
		const innerEnd = { x: center + innerRadius * Math.cos(endRad), y: center + innerRadius * Math.sin(endRad) };

		return [
			'M', outerStart.x, outerStart.y,
			'A', radius, radius, 0, largeArc, 1, outerEnd.x, outerEnd.y,
			'L', innerEnd.x, innerEnd.y,
			'A', innerRadius, innerRadius, 0, largeArc, 0, innerStart.x, innerStart.y,
			'Z'
		].join(' ');
	}

	// Tooltip
	let hoveredIndex = $state<number | null>(null);
	let tooltipX = $state(0);
	let tooltipY = $state(0);

	function handleMouseMove(e: MouseEvent) {
		const rect = (e.currentTarget as SVGElement).getBoundingClientRect();
		tooltipX = e.clientX - rect.left;
		tooltipY = e.clientY - rect.top - 10;
	}
</script>

<div class={className}>
	<div class="flex flex-col items-center">
		<div class="relative">
			<svg
				width={size}
				height={size}
				viewBox="0 0 {size} {size}"
				class="drop-shadow-sm"
				role="img"
				aria-label="Sector allocation chart"
				onmousemove={handleMouseMove}
				onmouseleave={() => (hoveredIndex = null)}
			>
				<g>
					{#each slices() as slice, i}
						<path
							d={slice.path}
							fill={slice.color}
							class="cursor-pointer transition-all duration-150 hover:opacity-90"
							stroke="white"
							stroke-width="1"
							role="button"
							tabindex="0"
							onmouseenter={() => (hoveredIndex = i)}
						/>
					{/each}
				</g>

				<!-- Center Text -->
				<text x={center} y={center} text-anchor="middle" dominant-baseline="middle" class="fill-foreground text-xs font-medium">
					{data.length} Sectors
				</text>
			</svg>

			{#if hoveredIndex !== null && slices()[hoveredIndex]}
				<div
					class="pointer-events-none absolute z-10 rounded-lg border border-border bg-popover px-3 py-2 text-xs shadow-lg"
					style="left: {tooltipX}px; top: {tooltipY}px; transform: translate(-50%, -100%);"
				>
					<p class="font-semibold" style="color: {slices()[hoveredIndex].color}">{slices()[hoveredIndex].label}</p>
					<p class="mt-1 tabular-nums">{slices()[hoveredIndex].percent.toFixed(1)}%</p>
					<p class="text-muted-foreground tabular-nums">Rs. {slices()[hoveredIndex].value.toLocaleString()}</p>
				</div>
			{/if}
		</div>

		<!-- Legend -->
		<div class="mt-4 grid grid-cols-2 gap-x-6 gap-y-2 text-xs">
			{#each slices() as slice}
				<div class="flex items-center gap-2">
					<div class="h-2.5 w-2.5 rounded-sm" style="background: {slice.color};"></div>
					<span class="truncate max-w-[100px]" title={slice.label}>{slice.label}</span>
					<span class="ml-auto font-medium tabular-nums">{slice.percent.toFixed(1)}%</span>
				</div>
			{/each}
		</div>
	</div>
</div>
