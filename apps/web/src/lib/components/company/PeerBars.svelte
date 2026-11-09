<script lang="ts">
	import type { Peer } from '#lib/gen/ntx/v1/story_pb.js';

	interface Props {
		peers: Peer[];
		metric: 'pe' | 'pb';
		symbol: string;
		median?: number; // the sector median for the metric
		period: string; // this company's newest filed period
		class?: string;
	}

	let { peers, metric, symbol, median, period, class: className = '' }: Props = $props();

	// Hydropower has over a hundred peers; the nearest few each way answer "where does it sit".
	const WINDOW = 5;
	let showAll = $state(false);
	let sorted = $derived(
		peers.filter((p) => p[metric] !== undefined).sort((a, b) => (a[metric] ?? 0) - (b[metric] ?? 0))
	);
	let windowed = $derived(sorted.length > WINDOW * 2 + 2);
	let shown = $derived.by(() => {
		if (showAll || !windowed) return sorted;
		const i = Math.max(
			0,
			sorted.findIndex((p) => p.symbol === symbol)
		);
		return sorted.slice(Math.max(0, i - WINDOW), i + WINDOW + 1);
	});
	let max = $derived(Math.max(...shown.map((p) => p[metric] ?? 0), median ?? 0) * 1.05);
	const fmt = (v: number) => `${v.toLocaleString('en-NP', { maximumFractionDigits: 2 })}x`;
	const GRID = 'grid grid-cols-[4.5rem_1fr_3.5rem] items-center gap-3';
</script>

{#if shown.length > 0}
	<div class={className}>
		{#if median !== undefined}
			<div class="{GRID} h-5 text-[11px] text-muted-foreground tabular-nums" aria-hidden="true">
				<span></span>
				<span class="relative block h-full">
					<span
						class="absolute -translate-x-1/2 whitespace-nowrap"
						style="left: {(median / max) * 100}%">median {fmt(median)}</span
					>
				</span>
				<span></span>
			</div>
		{/if}
		<ol class="space-y-1.5 text-sm tabular-nums">
			{#each shown as p (p.symbol)}
				{@const mine = p.symbol === symbol}
				<li class={GRID}>
					<a
						href="/company/{p.symbol}"
						class="truncate hover:underline {mine ? 'font-medium' : 'text-muted-foreground'}"
						title={p.name}>{p.symbol}</a
					>
					<span class="relative block h-3">
						<span
							class="absolute inset-y-0 left-0 rounded-sm {mine
								? 'bg-primary'
								: 'bg-foreground/25'}"
							style="width: {((p[metric] ?? 0) / max) * 100}%"
						></span>
						{#if median !== undefined}
							<!-- The median, drawn per row so the dashes read as one line down the chart. -->
							<span
								class="absolute -inset-y-1 border-l border-dashed border-foreground/60"
								style="left: {(median / max) * 100}%"
								aria-hidden="true"
							></span>
						{/if}
					</span>
					<span class="text-right {mine ? 'font-medium' : 'text-muted-foreground'}"
						>{fmt(p[metric] ?? 0)}</span
					>
				</li>
			{/each}
		</ol>
		<p class="mt-3 text-xs text-muted-foreground">
			{#if windowed}
				<button
					onclick={() => (showAll = !showAll)}
					aria-expanded={showAll}
					class="underline hover:text-foreground"
					>{showAll
						? `Show the nearest ${WINDOW} each way`
						: `Show all ${sorted.length} peers`}</button
				>
				·
			{/if}
			{metric.toUpperCase()} from each company's latest filed quarter and latest close; this company's
			is {period || 'not recorded'}.
		</p>
	</div>
{:else}
	<p class="text-sm text-muted-foreground {className}">
		No peer has a {metric.toUpperCase()} on file.
	</p>
{/if}
