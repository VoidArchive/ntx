<script lang="ts">
	import type { Plant } from '#lib/gen/ntx/v1/hydro_pb.js';
	import { STATUS_LABEL, num } from '#lib/hydro.js';
	import StateMark from './StateMark.svelte';

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
				<th class="py-2 text-left font-normal">Source</th>
			</tr>
		</thead>
		<tbody>
			{#each plants as p (p.id)}
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
					<td class="py-1.5"><StateMark state={p.state} stale={p.stale} /></td>
					<!-- Each row names its own reading day and sources (ADR 0002). -->
					<td class="py-1.5 text-xs text-muted-foreground">
						{#if p.latest}
							{p.latest.day} · {num(p.latest.rainMm)} mm · {num(p.latest.dischargeM3s)} m³/s
						{:else}
							not recorded
						{/if}
					</td>
					<!-- Each row names its own sources (ADR 0002): the licence row for the plant, the models for the day. -->
					<td class="py-1.5 text-xs leading-tight text-muted-foreground">
						<a href={p.sourceUrl} class="underline underline-offset-2">DoED licence row</a>
						{#if p.latest}
							<br />{p.latest.rainSource || 'rain not recorded'}<br />{p.latest.dischargeSource ||
								'discharge not recorded'}
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
	<p class="mt-2 text-xs text-muted-foreground">
		States follow the rules printed on <a href="/hydro" class="underline underline-offset-2"
			>the hydro page</a
		>.
	</p>
</div>
