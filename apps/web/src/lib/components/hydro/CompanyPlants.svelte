<script lang="ts">
	import type { Plant } from '#lib/gen/ntx/v1/hydro_pb.js';
	import { STATE, STATUS_LABEL, num } from '#lib/hydro.js';

	let { plants }: { plants: Plant[] } = $props();
</script>

<div class="overflow-x-auto">
	<table class="w-full text-sm whitespace-nowrap tabular-nums [&_td+td]:pl-3 [&_th+th]:pl-3">
		<thead class="text-xs text-muted-foreground">
			<tr class="border-b border-border">
				<th class="py-2 text-left font-normal">Plant</th>
				<th class="py-2 text-left font-normal">River</th>
				<th class="py-2 text-left font-normal">Basin</th>
				<th class="py-2 text-right font-normal">MW</th>
				<th class="py-2 text-left font-normal">Status</th>
				<th class="py-2 text-left font-normal">State</th>
				<th class="py-2 text-left font-normal">Reading</th>
			</tr>
		</thead>
		<tbody>
			{#each plants as p (p.id)}
				{@const s = STATE[p.state]}
				<tr class="border-b border-border/50">
					<td class="py-1.5 font-medium">{p.plant}</td>
					<td class="py-1.5 text-muted-foreground">{p.river}</td>
					<td class="py-1.5">
						<a
							href="/hydro?basin={encodeURIComponent(p.basin)}"
							class="underline underline-offset-2">{p.basin}</a
						>
					</td>
					<td class="py-1.5 text-right">{num(p.mw, 2)}</td>
					<td class="py-1.5 text-muted-foreground">{STATUS_LABEL[p.status] ?? p.status}</td>
					<td class="py-1.5">
						<span aria-hidden="true" style="color: {s.color}">{s.glyph}</span>
						{s.label}{#if p.stale}
							<span class="text-xs text-muted-foreground">(stale)</span>{/if}
					</td>
					<!-- Each row names its own reading day and sources (ADR 0002). -->
					<td class="py-1.5 text-xs text-muted-foreground">
						{#if p.latest}
							{p.latest.day} · {num(p.latest.rainMm)} mm · {num(p.latest.dischargeM3s)} m³/s
						{:else}
							not recorded
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
	<p class="mt-2 text-xs text-muted-foreground">
		Coordinates and capacity from the
		<a href={plants[0]?.sourceUrl} class="underline underline-offset-2">DoED licence table</a>; rain
		and discharge from {plants[0]?.latest?.rainSource || 'Open-Meteo, ECMWF IFS'} and
		{plants[0]?.latest?.dischargeSource || 'Open-Meteo, GloFAS v4'}. Rules on
		<a href="/hydro" class="underline underline-offset-2">the hydro page</a>.
	</p>
</div>
