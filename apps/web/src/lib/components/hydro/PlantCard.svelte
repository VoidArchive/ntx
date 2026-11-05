<script lang="ts">
	import X from '@lucide/svelte/icons/x';
	import type { Plant, PlantState } from '#lib/gen/ntx/v1/hydro_pb.js';
	import { STATUS_LABEL, num } from '#lib/hydro.js';
	import StateMark from './StateMark.svelte';

	// The readings route serves plain JSON, not proto messages.
	type Reading = { day: string; rainMm?: number; dischargeM3s?: number; state: PlantState };

	interface Props {
		plant: Plant;
		// Seven days of readings, newest first, or the fetch's state.
		readings?: Reading[] | 'loading' | 'failed';
		pinned?: boolean;
		onClose?: () => void;
	}

	let { plant: p, readings, pinned = false, onClose }: Props = $props();
</script>

<!-- One plant and its weather. Hover shows it; a click on the dot or the row pins it. -->
<div
	class="w-72 rounded-md border border-border bg-popover p-3 text-xs text-popover-foreground shadow-md"
>
	<div class="flex items-start justify-between gap-2">
		<div class="min-w-0">
			<p class="text-sm font-medium">{p.plant}</p>
			<p class="truncate text-muted-foreground" title={p.companyName}>
				<a href="/company/{p.symbol}" class="font-serif font-medium text-foreground hover:underline"
					>{p.symbol}</a
				>
				· {p.companyName}
			</p>
		</div>
		{#if pinned}
			<button
				onclick={onClose}
				aria-label="Close"
				class="-m-1 shrink-0 rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
			>
				<X class="size-3.5" />
			</button>
		{/if}
	</div>

	<p class="mt-2 tabular-nums">
		{p.river}, {p.basin} · {num(p.mw, 2)} MW · {STATUS_LABEL[p.status] ?? p.status}
	</p>
	<p class="mt-1">
		<StateMark state={p.state} stale={p.stale} />
		{#if p.latest}
			<span class="text-muted-foreground tabular-nums">
				· {p.latest.day}: {num(p.latest.rainMm)} mm rain, {num(p.latest.dischargeM3s)} m³/s
			</span>
		{:else}
			<span class="text-muted-foreground">· no readings recorded</span>
		{/if}
	</p>
	{#if p.threshold}
		<p class="mt-1 text-muted-foreground">{p.threshold}</p>
	{/if}

	<!-- Last seven days, newest first -->
	<div class="mt-2 border-t border-border pt-2">
		{#if readings === undefined || readings === 'loading'}
			<p class="text-muted-foreground">Loading the last seven days…</p>
		{:else if readings === 'failed'}
			<p class="text-muted-foreground">Readings could not be loaded.</p>
		{:else if readings.length === 0}
			<p class="text-muted-foreground">No readings recorded for this plant.</p>
		{:else}
			<table class="w-full tabular-nums [&_td+td]:pl-2 [&_th+th]:pl-2">
				<thead class="text-muted-foreground">
					<tr>
						<th class="text-left font-normal">Day</th>
						<th class="text-right font-normal">Rain, mm</th>
						<th class="text-right font-normal">m³/s</th>
						<th class="text-left font-normal">State</th>
					</tr>
				</thead>
				<tbody>
					{#each readings as r (r.day)}
						<tr>
							<td>{r.day.slice(5)}</td>
							<td class="text-right">{num(r.rainMm)}</td>
							<td class="text-right">{num(r.dischargeM3s)}</td>
							<td><StateMark state={r.state} /></td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</div>

	<p class="mt-2 border-t border-border pt-2 text-muted-foreground tabular-nums">
		Close {num(p.close, 2)} · PE {p.pe === undefined ? '—' : `${num(p.pe)}x`} · rating {p.rating ||
			'—'}
		{#if p.period}· {p.period}{/if}
	</p>
	{#if !pinned}
		<p class="mt-1 text-muted-foreground">Click the dot to keep this open.</p>
	{/if}
</div>
