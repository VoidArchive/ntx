<script lang="ts">
	import { enhance } from '$app/forms';
	import { GateKind, type Gate } from '#lib/gen/ntx/v1/portfolio_pb.js';
	import { VERDICT, byVerdict, directions, figure, isEvent, metrics } from './gates.js';

	// One holding's gates with their verdicts, a remove per row and the entry form.
	let { symbol, gates }: { symbol: string; gates: Gate[] } = $props();

	const field = 'h-8 rounded-md border border-input bg-background px-2 text-sm';
	const button = 'h-8 rounded-md border border-input bg-background px-3 text-sm hover:bg-accent';
	const today = new Date().toISOString().slice(0, 10);

	let kind = $state(GateKind.METRIC);
	let sorted = $derived([...gates].sort(byVerdict));
</script>

<h3 class="text-xs text-muted-foreground">Gates</h3>
<table class="mt-1 w-full text-sm tabular-nums [&_td+td]:pl-3">
	<tbody>
		{#each sorted as g (g.id)}
			<tr class="border-b border-border/50">
				<td class="py-1">
					{g.label}
					{#if isEvent(g)}
						<span class="text-xs text-muted-foreground">· checked {g.checkedOn || 'never'}</span>
					{/if}
				</td>
				<td class="py-1 text-right">
					{#if !isEvent(g)}
						{figure(g.value, g.unit)}{#if g.prior !== undefined}
							<span class="text-xs text-muted-foreground"> vs {figure(g.prior, g.unit)}</span>{/if}
					{/if}
				</td>
				<td class="py-1 text-xs text-muted-foreground">
					{#if !isEvent(g)}
						{#if g.source}<a href={g.source} class="underline" rel="noreferrer">{g.period}</a
							>{:else}{g.period || 'no period'}{/if}
					{/if}
				</td>
				<td class="py-1 font-medium">{VERDICT[g.verdict] ?? ''}</td>
				<td class="py-1 text-right">
					<form method="POST" action="?/deleteGate" use:enhance>
						<input type="hidden" name="id" value={g.id} />
						<button type="submit" class="text-xs text-muted-foreground underline underline-offset-2"
							>Remove</button
						>
					</form>
				</td>
			</tr>
		{:else}
			<tr><td class="py-1 text-muted-foreground">No gates written.</td></tr>
		{/each}
	</tbody>
</table>
<form
	method="POST"
	action="?/gate"
	use:enhance
	class="mt-3 flex flex-wrap items-center gap-2 text-sm"
>
	<input type="hidden" name="symbol" value={symbol} />
	<select name="kind" bind:value={kind} class={field} aria-label="Gate kind">
		<option value={GateKind.METRIC}>Metric</option>
		<option value={GateKind.EVENT}>Event</option>
	</select>
	{#if kind === GateKind.METRIC}
		<select name="metric" class={field} aria-label="Metric">
			{#each metrics as m (m.value)}<option value={m.value}>{m.label}</option>{/each}
		</select>
		<select name="direction" class={field} aria-label="Direction">
			{#each directions as d (d.value)}<option value={d.value}>{d.label}</option>{/each}
		</select>
		<input
			type="number"
			name="threshold"
			step="any"
			aria-label="Threshold"
			placeholder="threshold"
			class="{field} w-28"
		/>
		<input
			type="number"
			name="periods"
			min="1"
			step="1"
			value="1"
			aria-label="Consecutive periods"
			class="{field} w-20"
		/>
	{:else}
		<input
			type="text"
			name="text"
			required
			aria-label="Event"
			placeholder="what would break the thesis"
			class="{field} min-w-64 flex-1"
		/>
		<input type="date" name="checked_on" value={today} class={field} aria-label="Checked on" />
	{/if}
	<button type="submit" class={button}>Write</button>
</form>
