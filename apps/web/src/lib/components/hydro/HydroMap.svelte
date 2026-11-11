<script lang="ts">
	import type { Snippet } from 'svelte';
	import { MediaQuery } from 'svelte/reactivity';
	import type { Basin, Plant } from '#lib/gen/ntx/v1/hydro_pb.js';
	import { PlantState } from '#lib/gen/ntx/v1/hydro_pb.js';
	import map from '#lib/data/hydro_map.json';
	import { STATE, STATES } from '#lib/hydro.js';

	interface Props {
		plants: Plant[];
		basins: Basin[];
		basinFilter?: string;
		onBasin?: (name: string) => void;
		onHover?: (plant: Plant | undefined) => void;
		// The pinned plant: its dot is ringed, the others recede, and its card stays open.
		selected?: Plant;
		onSelect?: (plant: Plant | undefined) => void;
		// The card beside the active dot; the page owns its data.
		card?: Snippet<[Plant, boolean]>;
		hovered?: Plant;
	}

	let {
		plants,
		basins,
		basinFilter = '',
		onBasin,
		onHover,
		selected,
		onSelect,
		card,
		hovered
	}: Props = $props();

	// Tailwind's lg: the sidebar's breakpoint, and the map's hover-or-tap one.
	const desktop = new MediaQuery('(min-width: 64rem)');

	// The same Mercator the build script used, so dots land on the rivers it drew (ADR 0004).
	const { k, lon0, lat0, tx, ty, width, height } = map.projection;
	const rad = (deg: number) => (deg * Math.PI) / 180;
	const merc = (lat: number) => Math.log(Math.tan(Math.PI / 4 + rad(lat) / 2));
	function project(lon: number, lat: number): [number, number] {
		return [rad(lon - lon0) * k + tx, -(merc(lat) - merc(lat0)) * k + ty];
	}

	let basinState = $derived(new Map(basins.map((b) => [b.name, b.state])));
	const rivers = Object.groupBy(map.rivers, (r) => r.basin);

	// Square-root scale so a 456 MW plant is big without hiding its neighbours.
	const radius = (mw: number) => 1.5 + Math.sqrt(mw) * 0.8;
	const strokeWidth = (order: number) =>
		order >= 6 ? 1.4 : order === 5 ? 1 : order === 4 ? 0.7 : 0.45;
	const dim = (basin: string) => (basinFilter && basinFilter !== basin ? 0.25 : 1);
	const stateColor = (s: PlantState) => STATE[s].color;
	const dots = $derived([...plants].sort((a, b) => b.mw - a.mw));
	// A tap leaves a phone's emulated hover behind, so only a mouse's hover counts there.
	let active = $derived(selected ?? (desktop.current ? hovered : undefined));
	const dotOpacity = (p: Plant) =>
		Math.min(dim(p.basin), selected && selected.id !== p.id ? 0.35 : 1);
	// The tap target: a finger needs more than a 5 MW dot, a mouse in a cluster needs less.
	const hit = (mw: number) => Math.max(radius(mw), desktop.current ? 5 : 12);

	// On a phone the basin pick is also the zoom: the view box closes on the basin's plants.
	let view = $derived.by(() => {
		const inBasin = desktop.current ? [] : plants.filter((p) => p.basin === basinFilter);
		if (inBasin.length === 0) return `0 0 ${width} ${height}`;
		const pts = inBasin.map((p) => project(p.lon, p.lat));
		const pad = 40;
		const x0 = Math.min(...pts.map(([x]) => x)) - pad;
		const y0 = Math.min(...pts.map(([, y]) => y)) - pad;
		let w = Math.max(...pts.map(([x]) => x)) + pad - x0;
		let h = Math.max(...pts.map(([, y]) => y)) + pad - y0;
		// Keep the map's aspect so the box fills the width without letterboxing.
		const [bw, bh] = [w, h];
		if (w / h > width / height) h = (w * height) / width;
		else w = (h * width) / height;
		return `${x0 - (w - bw) / 2} ${y0 - (h - bh) / 2} ${w} ${h}`;
	});

	function pick(basin: string) {
		onBasin?.(basinFilter === basin ? '' : basin);
	}

	// The card sits beside the dot, on whichever side has room, and never past the map's edge.
	function cardStyle(p: Plant): string {
		const [x, y] = project(p.lon, p.lat);
		const dx = x > width / 2 ? 'calc(-100% - 12px)' : '12px';
		const dy = y < height * 0.35 ? '0' : y > height * 0.65 ? '-100%' : '-50%';
		return `left: ${(x / width) * 100}%; top: ${(y / height) * 100}%; transform: translate(${dx}, ${dy})`;
	}
</script>

<svelte:window onkeydown={(e) => e.key === 'Escape' && selected && onSelect?.(undefined)} />

<figure class="w-full" aria-label="Map of Nepal's rivers and listed hydro plants">
	<!-- Bleeds to the card's edge on a phone, as the tables do. -->
	<div id="hydro-map" class="relative -mx-6 sm:mx-0">
		<svg viewBox={view} class="h-auto w-full" role="img" aria-describedby="hydro-map-legend">
			<path
				d={map.outline}
				fill="var(--color-muted)"
				fill-opacity="0.35"
				stroke="var(--color-border)"
				vector-effect="non-scaling-stroke"
			/>
			{#each Object.entries(rivers) as [basin, reaches] (basin)}
				<!-- Handlers sit on the basin group, not on each of its reaches: a click anywhere on a river is a basin pick. -->
				<g
					role="button"
					tabindex="0"
					aria-label="Filter by {basin}"
					opacity={dim(basin)}
					stroke={stateColor(basinState.get(basin) ?? PlantState.UNSPECIFIED)}
					fill="none"
					stroke-linecap="round"
					stroke-linejoin="round"
					onclick={() => pick(basin)}
					onkeydown={(e) => e.key === 'Enter' && pick(basin)}
					class="cursor-pointer"
				>
					{#each reaches ?? [] as r, i (i)}
						<path d={r.d} stroke-width={strokeWidth(r.order)} vector-effect="non-scaling-stroke" />
					{/each}
				</g>
			{/each}
			{#each dots as p (p.id)}
				{@const [x, y] = project(p.lon, p.lat)}
				{@const operating = p.status === 'operating'}
				<circle
					cx={x}
					cy={y}
					r={radius(p.mw)}
					fill={operating ? stateColor(p.state) : 'var(--color-background)'}
					stroke={operating ? 'var(--color-background)' : stateColor(p.state)}
					stroke-width={operating ? 0.8 : 1.5}
					opacity={dotOpacity(p)}
					vector-effect="non-scaling-stroke"
					pointer-events="none"
				/>
			{/each}
			<!-- The targets sit above every dot, smallest plant last so it wins its cluster. -->
			{#each dots as p (p.id)}
				{@const [x, y] = project(p.lon, p.lat)}
				<circle
					cx={x}
					cy={y}
					r={hit(p.mw)}
					fill="transparent"
					class="cursor-pointer focus:outline-none"
					role="button"
					tabindex="0"
					aria-label="{p.plant}, {p.symbol}, {p.mw} MW, {STATE[p.state].label}"
					aria-pressed={selected?.id === p.id}
					onmouseenter={() => onHover?.(p)}
					onmouseleave={() => onHover?.(undefined)}
					onfocus={() => onHover?.(p)}
					onblur={() => onHover?.(undefined)}
					onclick={() => onSelect?.(selected?.id === p.id ? undefined : p)}
					onkeydown={(e) => e.key === 'Enter' && onSelect?.(p)}
				/>
			{/each}
			{#if active}
				<!-- The ring marks the active plant above every neighbour; the focus ring is this, not the browser's. -->
				{@const [x, y] = project(active.lon, active.lat)}
				<circle
					cx={x}
					cy={y}
					r={radius(active.mw) + 3}
					fill="none"
					stroke="var(--color-foreground)"
					stroke-width="1.5"
					vector-effect="non-scaling-stroke"
					pointer-events="none"
				/>
			{/if}
		</svg>
		{#if active && card}
			<div
				class="absolute z-10 hidden lg:block {selected ? '' : 'pointer-events-none'}"
				style={cardStyle(active)}
				aria-live="polite"
			>
				{@render card(active, !!selected)}
			</div>
		{/if}
	</div>
	<!-- A phone has no hover, so only the pinned plant shows, as a block under the map. -->
	{#if selected && card}
		<div class="mt-3 lg:hidden" aria-live="polite">
			{@render card(selected, true)}
		</div>
	{/if}
	<figcaption
		id="hydro-map-legend"
		class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground"
	>
		{#each STATES as s (s)}
			<span
				><span aria-hidden="true" style="color: {STATE[s].color}">{STATE[s].glyph}</span>
				{STATE[s].label}</span
			>
		{/each}
		<span>● operating</span>
		<span>◌ under construction</span>
		<span>dot area follows MW</span>
		<span>rivers coloured by basin state</span>
	</figcaption>
</figure>
