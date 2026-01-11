<script lang="ts">
	interface DataPoint {
		label: string;
		value: number;
		max: number;
		sectorAvg?: number;
	}

	interface Props {
		data: DataPoint[];
		class?: string;
	}

	let { data, class: className = '' }: Props = $props();

	const size = 280;
	const center = size / 2;
	const radius = 100;
	const levels = 4;

	// Calculate points for a value on each axis
	function getPoint(index: number, value: number, max: number): { x: number; y: number } {
		const angle = (Math.PI * 2 * index) / data.length - Math.PI / 2;
		const normalizedValue = Math.min(value / max, 1);
		const r = radius * normalizedValue;
		return {
			x: center + r * Math.cos(angle),
			y: center + r * Math.sin(angle)
		};
	}

	// Generate polygon path from values
	function getPolygonPath(values: { value: number; max: number }[]): string {
		return (
			values
				.map((v, i) => {
					const point = getPoint(i, v.value, v.max);
					return `${i === 0 ? 'M' : 'L'} ${point.x} ${point.y}`;
				})
				.join(' ') + ' Z'
		);
	}

	// Company polygon
	let companyPath = $derived(getPolygonPath(data.map((d) => ({ value: d.value, max: d.max }))));

	// Sector average polygon (if available)
	let sectorPath = $derived.by(() => {
		const hasSomeSectorAvg = data.some((d) => d.sectorAvg !== undefined);
		if (!hasSomeSectorAvg) return null;
		// Use sectorAvg if available, otherwise use 0 to keep the polygon shape consistent
		return getPolygonPath(data.map((d) => ({ value: d.sectorAvg ?? 0, max: d.max })));
	});

	// Grid lines (concentric polygons)
	function getGridPath(level: number): string {
		const points = data.map((_d, i) => {
			const angle = (Math.PI * 2 * i) / data.length - Math.PI / 2;
			const r = (radius * level) / levels;
			return { x: center + r * Math.cos(angle), y: center + r * Math.sin(angle) };
		});
		return points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ') + ' Z';
	}

	// Axis lines from center to each vertex
	let axisLines = $derived(
		data.map((_, i) => {
			const angle = (Math.PI * 2 * i) / data.length - Math.PI / 2;
			return {
				x2: center + radius * Math.cos(angle),
				y2: center + radius * Math.sin(angle)
			};
		})
	);

	// Label positions (slightly outside the chart)
	let labelPositions = $derived(
		data.map((d, i) => {
			const angle = (Math.PI * 2 * i) / data.length - Math.PI / 2;
			const labelRadius = radius + 24;
			return {
				x: center + labelRadius * Math.cos(angle),
				y: center + labelRadius * Math.sin(angle),
				label: d.label,
				anchor: Math.abs(Math.cos(angle)) < 0.1 ? 'middle' : Math.cos(angle) > 0 ? 'start' : 'end'
			};
		})
	);
</script>

<div class="flex flex-col items-center {className}">
	<svg width={size} height={size} class="overflow-visible">
		<!-- Grid levels -->
		{#each Array(levels) as _, i (i)}
			<path
				d={getGridPath(i + 1)}
				fill="none"
				stroke="currentColor"
				stroke-opacity="0.1"
				stroke-width="1"
			/>
		{/each}

		<!-- Axis lines -->
		{#each axisLines as axis, i (i)}
			<line
				x1={center}
				y1={center}
				x2={axis.x2}
				y2={axis.y2}
				stroke="currentColor"
				stroke-opacity="0.15"
				stroke-width="1"
			/>
		{/each}

		<!-- Sector average polygon (if available) -->
		{#if sectorPath}
			<path
				d={sectorPath}
				fill="none"
				stroke="hsl(var(--muted-foreground))"
				stroke-opacity="0.4"
				stroke-width="2"
				stroke-dasharray="5 5"
			/>
		{/if}

		<!-- Company polygon -->
		<path
			d={companyPath}
			fill="hsl(25 95% 53% / 0.15)"
			stroke="hsl(25 95% 53%)"
			stroke-width="2.5"
		/>

		<!-- Data points -->
		{#each data as d, i (d.label)}
			{@const point = getPoint(i, d.value, d.max)}
			<circle cx={point.x} cy={point.y} r="4.5" fill="hsl(25 95% 53%)" />
		{/each}

		<!-- Labels -->
		{#each labelPositions as pos (pos.label)}
			<text
				x={pos.x}
				y={pos.y}
				text-anchor={pos.anchor}
				dominant-baseline="middle"
				class="fill-muted-foreground text-xs"
			>
				{pos.label}
			</text>
		{/each}
	</svg>

	<!-- Legend -->
	<div class="mt-4 flex items-center gap-6 text-xs">
		<div class="flex items-center gap-2">
			<div class="h-0.5 w-4 rounded" style="background: hsl(25 95% 53%);"></div>
			<span class="text-muted-foreground">Company</span>
		</div>
		{#if sectorPath}
			<div class="flex items-center gap-2">
				<div class="h-0.5 w-4 border-t-2 border-dashed border-muted-foreground/40"></div>
				<span class="text-muted-foreground">Sector Avg</span>
			</div>
		{/if}
	</div>
</div>
