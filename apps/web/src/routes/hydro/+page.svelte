<script lang="ts">
	import type { PageData } from './$types';
	import type { Plant } from '#lib/gen/ntx/v1/hydro_pb.js';
	import { PlantState } from '#lib/gen/ntx/v1/hydro_pb.js';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { SvelteMap, SvelteURLSearchParams } from 'svelte/reactivity';
	import HydroMap from '#lib/components/hydro/HydroMap.svelte';
	import StateMark from '#lib/components/hydro/StateMark.svelte';
	import { STATE, STATES, STATUS_LABEL, num, stateFromKey } from '#lib/hydro.js';

	let { data }: { data: PageData } = $props();

	type Reading = {
		day: string;
		rainMm?: number;
		dischargeM3s?: number;
		rainSource: string;
		dischargeSource: string;
		state: PlantState;
	};

	type SortKey = 'basin' | 'mw' | 'state' | 'pe';
	const columns: { key: SortKey; label: string; desc: boolean }[] = [
		{ key: 'basin', label: 'Basin', desc: false },
		{ key: 'mw', label: 'MW', desc: true },
		{ key: 'state', label: 'State', desc: true },
		{ key: 'pe', label: 'PE', desc: false }
	];

	// The URL is the only filter state, as on the screener, so a basin and a PE sort are a link.
	// A shallow goto leaves page.url on the loaded URL and puts the target in page.shallow.
	let url = $derived(page.shallow?.url ?? page.url);
	let params = $derived(url.searchParams);
	let basinFilter = $derived(params.get('basin') ?? '');
	let statusFilter = $derived(params.get('status') ?? '');
	let stateFilter = $derived(stateFromKey(params.get('state') ?? ''));
	let sortCol = $derived(columns.find((c) => c.key === params.get('sort')) ?? columns[0]);
	let sortDesc = $derived(params.has('dir') ? params.get('dir') === 'desc' : sortCol.desc);
	let hasFilters = $derived(!!(basinFilter || statusFilter || stateFilter !== undefined));

	function setParams(changes: Record<string, string>) {
		const next = new SvelteURLSearchParams(url.searchParams.toString());
		for (const [k, v] of Object.entries(changes)) {
			if (v) next.set(k, v);
			if (!v) next.delete(k);
		}
		goto(`?${next}`, { replace: true, shallow: true });
	}

	function sortBy(col: (typeof columns)[number]) {
		const desc = col.key === sortCol.key ? !sortDesc : col.desc;
		setParams({ sort: col.key, dir: desc ? 'desc' : 'asc' });
	}

	let rows = $derived.by(() => {
		const out = data.plants.filter((p) => {
			if (basinFilter && p.basin !== basinFilter) return false;
			if (statusFilter && p.status !== statusFilter) return false;
			if (stateFilter !== undefined && p.state !== stateFilter) return false;
			return true;
		});
		const dir = sortDesc ? -1 : 1;
		return out.sort((a, b) => {
			if (sortCol.key === 'basin')
				return dir * (a.basin.localeCompare(b.basin) || a.river.localeCompare(b.river));
			if (sortCol.key === 'mw') return dir * (a.mw - b.mw);
			if (sortCol.key === 'state') return dir * (a.state - b.state) || b.mw - a.mw;
			// Missing PE sorts last whichever way the column points.
			if (a.pe === undefined || b.pe === undefined)
				return Number(a.pe === undefined) - Number(b.pe === undefined);
			return dir * (a.pe - b.pe);
		});
	});

	let hovered = $state<Plant | undefined>();
	const open = new SvelteMap<bigint, Reading[] | 'loading' | 'failed'>();

	async function toggle(p: Plant) {
		if (open.has(p.id)) {
			open.delete(p.id);
			return;
		}
		open.set(p.id, 'loading');
		try {
			const res = await fetch(`/hydro/readings/${p.id}`);
			if (!res.ok) throw new Error(res.statusText);
			open.set(p.id, (await res.json()) as Reading[]);
		} catch {
			open.set(p.id, 'failed');
		}
	}

	let basinsWithPlants = $derived(data.basins.filter((b) => b.plantCount > 0));
	let lastRun = $derived(data.lastRun ? `${data.lastRun} UTC` : 'not yet');
	const thClass = 'py-2 font-normal';
</script>

<svelte:head>
	<title>Hydro · NTX</title>
</svelte:head>

<div class="mx-auto min-h-screen max-w-7xl px-4 py-8">
	<header>
		<h1 class="font-serif text-3xl font-medium tracking-tight">Hydro</h1>
		<!-- 1. The sentence first, as on every page. -->
		<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">{data.opening}</p>
	</header>

	<!-- Basin states double as the basin filter; on a phone they stand in for the map. -->
	<ul class="mt-4 flex flex-wrap gap-1.5 text-xs" aria-label="Basins">
		{#each basinsWithPlants as b (b.name)}
			{@const s = STATE[b.state]}
			<li>
				<button
					onclick={() => setParams({ basin: basinFilter === b.name ? '' : b.name })}
					aria-pressed={basinFilter === b.name}
					class="rounded-full px-2.5 py-1 transition-colors {basinFilter === b.name
						? 'bg-foreground text-background'
						: 'bg-muted text-muted-foreground hover:bg-accent hover:text-foreground'}"
				>
					<span aria-hidden="true" style={basinFilter === b.name ? '' : `color: ${s.color}`}
						>{s.glyph}</span
					>
					{b.name} · {s.label} · {b.plantCount}
				</button>
			</li>
		{/each}
	</ul>

	<div class="mt-6 flex flex-col gap-8">
		<!-- 2. Map: desktop above the list with hover; phone below it, read-only. -->
		<section class="order-3 lg:order-1" aria-label="Map">
			<div class="hidden lg:block">
				<HydroMap
					plants={data.plants}
					basins={data.basins}
					{basinFilter}
					onBasin={(name) => setParams({ basin: name })}
					onHover={(p) => (hovered = p)}
				/>
				<p class="mt-2 min-h-5 text-sm text-muted-foreground tabular-nums" aria-live="polite">
					{#if hovered}
						<span class="font-medium text-foreground">{hovered.plant}</span>
						·
						<a href="/company/{hovered.symbol}" class="underline underline-offset-2"
							>{hovered.symbol}</a
						>
						· {hovered.river}, {hovered.basin} · {num(hovered.mw, 2)} MW · {STATUS_LABEL[
							hovered.status
						]}
						· <StateMark state={hovered.state} stale={hovered.stale} />
						{#if hovered.latest}
							· {hovered.latest.day}: {num(hovered.latest.rainMm)} mm, {num(
								hovered.latest.dischargeM3s
							)} m³/s
						{/if}
					{:else}
						Hover a dot for the plant; click a river or dot to filter the list by its basin.
					{/if}
				</p>
			</div>
			<div class="lg:hidden">
				<HydroMap plants={data.plants} basins={data.basins} {basinFilter} compact />
			</div>
		</section>

		<!-- 3. Plant list: the phone and keyboard surface. -->
		<section class="order-2" aria-label="Plants">
			<div class="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm">
				<label class="flex items-center gap-1">
					Status
					<select
						value={statusFilter}
						onchange={(e) => setParams({ status: e.currentTarget.value })}
						class="h-8 rounded-md border border-input bg-background px-2 text-sm"
					>
						<option value="">any</option>
						<option value="operating">operating</option>
						<option value="under_construction">under construction</option>
					</select>
				</label>
				<label class="flex items-center gap-1">
					State
					<select
						value={stateFilter === undefined ? '' : STATE[stateFilter].key}
						onchange={(e) => setParams({ state: e.currentTarget.value })}
						class="h-8 rounded-md border border-input bg-background px-2 text-sm"
					>
						<option value="">any</option>
						{#each STATES as s (s)}
							<option value={STATE[s].key}>{STATE[s].label}</option>
						{/each}
					</select>
				</label>
				{#if hasFilters}
					<button
						onclick={() => goto('?', { replace: true, shallow: true })}
						class="text-muted-foreground underline underline-offset-2 hover:text-foreground"
						>Clear</button
					>
				{/if}
				<span class="text-xs text-muted-foreground tabular-nums"
					>{rows.length} of {data.plants.length} plants</span
				>
			</div>

			<div class="mt-2 overflow-x-auto">
				<table class="w-full text-sm whitespace-nowrap tabular-nums [&_td+td]:pl-3 [&_th+th]:pl-3">
					<thead class="text-xs text-muted-foreground">
						<tr class="border-b border-border">
							<th class="{thClass} text-left"><span class="sr-only">Readings</span></th>
							<th class="{thClass} text-left">Plant</th>
							<th class="{thClass} text-left">Symbol</th>
							<th class="{thClass} text-left">River</th>
							{#each columns as col (col.key)}
								<th
									class="{thClass} {col.key === 'basin' ? 'text-left' : 'text-right'}"
									aria-sort={col.key === sortCol.key
										? sortDesc
											? 'descending'
											: 'ascending'
										: undefined}
								>
									<button
										onclick={() => sortBy(col)}
										class="hover:text-foreground {col.key === sortCol.key ? 'text-foreground' : ''}"
									>
										{col.label}{#if col.key === sortCol.key}
											<span aria-hidden="true">{sortDesc ? ' ▼' : ' ▲'}</span>{/if}
									</button>
								</th>
								{#if col.key === 'mw'}
									<th class="{thClass} text-right">Rain · discharge</th>
								{/if}
								{#if col.key === 'state'}
									<th class="{thClass} text-left">Reading</th>
									<th class="{thClass} text-right">Close</th>
								{/if}
								{#if col.key === 'pe'}
									<th class="{thClass} text-left">Rating</th>
									<th class="{thClass} text-left">As of</th>
								{/if}
							{/each}
						</tr>
					</thead>
					<tbody>
						{#each rows as p (p.id)}
							{@const readings = open.get(p.id)}
							<tr
								class="border-b border-border/50 hover:bg-muted/50 {hovered?.id === p.id
									? 'bg-muted/50'
									: ''}"
							>
								<td class="py-1.5">
									<button
										onclick={() => toggle(p)}
										aria-expanded={open.has(p.id)}
										aria-label="Last seven days for {p.plant}"
										class="w-5 text-muted-foreground hover:text-foreground"
										>{open.has(p.id) ? '▾' : '▸'}</button
									>
								</td>
								<td class="max-w-56 truncate py-1.5 font-medium" title={p.plant}>
									{p.plant}{#if p.status !== 'operating'}
										<span class="text-xs font-normal text-muted-foreground">
											({STATUS_LABEL[p.status]})</span
										>{/if}
								</td>
								<td class="py-1.5">
									<a
										href="/company/{p.symbol}"
										class="font-serif font-medium hover:underline"
										title={p.companyName}>{p.symbol}</a
									>
								</td>
								<td class="py-1.5 text-muted-foreground">{p.river}</td>
								<td class="py-1.5">{p.basin}</td>
								<td class="py-1.5 text-right">{num(p.mw, 2)}</td>
								<td class="py-1.5 text-right">
									{#if p.latest}
										{num(p.latest.rainMm)} mm · {num(p.latest.dischargeM3s)} m³/s
									{:else}
										not recorded
									{/if}
								</td>
								<td class="py-1.5 text-right"><StateMark state={p.state} stale={p.stale} /></td>
								<td class="py-1.5 text-xs text-muted-foreground"
									>{p.latest?.day ?? 'not recorded'}</td
								>
								<td class="py-1.5 text-right">{num(p.close, 2)}</td>
								<td class="py-1.5 text-right">{p.pe === undefined ? '—' : `${num(p.pe)}x`}</td>
								<td class="py-1.5">{p.rating || '—'}</td>
								<!-- Close, PE and rating name their own filing and close (ADR 0002). -->
								<td class="py-1.5 text-xs text-muted-foreground"
									>{p.period || 'not recorded'} · {p.closeDate || 'not recorded'}</td
								>
							</tr>
							{#if readings}
								<tr class="border-b border-border/50 bg-muted/20">
									<td></td>
									<td colspan="12" class="py-2 pr-3">
										{#if readings === 'loading'}
											<span class="text-xs text-muted-foreground">Loading…</span>
										{:else if readings === 'failed'}
											<span class="text-xs text-muted-foreground"
												>Readings could not be loaded.</span
											>
										{:else if readings.length === 0}
											<span class="text-xs text-muted-foreground"
												>No readings recorded for this plant.</span
											>
										{:else}
											<!-- 4. Seven days, newest first, a source beside each figure. -->
											<table class="text-xs [&_td+td]:pl-4 [&_th+th]:pl-4">
												<thead class="text-muted-foreground">
													<tr>
														<th class="text-left font-normal">Day</th>
														<th class="text-right font-normal">Rain, mm</th>
														<th class="text-left font-normal">Source</th>
														<th class="text-right font-normal">Discharge, m³/s</th>
														<th class="text-left font-normal">Source</th>
														<th class="text-left font-normal">State</th>
													</tr>
												</thead>
												<tbody>
													{#each readings as r (r.day)}
														<tr>
															<td>{r.day}</td>
															<td class="text-right">{num(r.rainMm)}</td>
															<td class="text-muted-foreground">{r.rainSource || 'not recorded'}</td
															>
															<td class="text-right">{num(r.dischargeM3s)}</td>
															<td class="text-muted-foreground"
																>{r.dischargeSource || 'not recorded'}</td
															>
															<td><StateMark state={r.state} /></td>
														</tr>
													{/each}
												</tbody>
											</table>
											{#if p.threshold}
												<p class="mt-1 text-xs text-muted-foreground">{p.threshold}</p>
											{/if}
										{/if}
									</td>
								</tr>
							{/if}
						{:else}
							<tr>
								<td colspan="13" class="py-12 text-center text-muted-foreground">
									No plant matches.
									<button
										onclick={() => goto('?', { replace: true, shallow: true })}
										class="underline">Clear filters</button
									>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</section>
	</div>

	<!-- 5. The rules, as the API states them, and where every figure comes from. -->
	<footer class="mt-12 border-t border-border pt-6 text-xs text-muted-foreground">
		<h2 class="text-xs tracking-wide uppercase">Rules and data</h2>
		<ul class="mt-2 space-y-1">
			{#each data.rules as rule (rule.name)}
				<li><span class="font-medium text-foreground">{rule.name}.</span> {rule.text}</li>
			{/each}
		</ul>
		<p class="mt-3 tabular-nums">
			Last weather run {lastRun} · plants from the
			<a
				href="https://www.doed.gov.np/pages/powerplantsmorethan1"
				class="underline underline-offset-2">DoED licence tables</a
			>, reviewed by hand, one source URL per row · weather by
			<a href="https://open-meteo.com/" class="underline underline-offset-2">Open-Meteo</a>
			(CC BY 4.0) · outline by
			<a href="https://www.geoboundaries.org" class="underline underline-offset-2">geoBoundaries</a>
			(CC BY 4.0) · rivers from
			<a href="https://www.hydrosheds.org" class="underline underline-offset-2"
				>HydroRIVERS, HydroSHEDS</a
			>
			· rain threshold from the
			<a href="https://www.dhm.gov.np" class="underline underline-offset-2"
				>Department of Hydrology and Meteorology</a
			>.
		</p>
	</footer>
</div>
