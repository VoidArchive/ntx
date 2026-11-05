<script lang="ts">
	import type { Peer, Range } from '#lib/gen/ntx/v1/story_pb.js';
	import { at } from '#lib/utils.js';

	interface Props {
		peers: Peer[];
		// The PE strip's sector range, so a peer sits where it would on that strip.
		range?: Range;
		symbol: string;
		// This company's newest filed period.
		period: string;
	}

	let { peers, range, symbol, period }: Props = $props();

	// Hydropower has over a hundred peers; the nearest few by PE answer "where does it sit".
	const WINDOW = 5;
	let showAll = $state(false);
	let windowed = $derived(peers.length > WINDOW * 2 + 2);
	let shown = $derived.by(() => {
		if (showAll || !windowed) return peers;
		const i = Math.max(
			0,
			peers.findIndex((p) => p.symbol === symbol)
		);
		return peers.slice(Math.max(0, i - WINDOW), i + WINDOW + 1);
	});

	function fmt(value: number | undefined): string {
		return value === undefined
			? '—'
			: `${value.toLocaleString('en-NP', { maximumFractionDigits: 2 })}x`;
	}
</script>

<div id="peers" class="mt-8">
	<h3 class="mb-4 font-serif text-base font-medium">Peers by PE</h3>
	<div class="overflow-x-auto">
		<table class="w-full text-sm whitespace-nowrap tabular-nums">
			<thead class="text-xs text-muted-foreground">
				<tr class="border-b border-border">
					<th class="py-2 text-left font-normal">Symbol</th>
					<th class="py-2 text-right font-normal">PE</th>
					<th class="py-2 text-right font-normal">PB</th>
					<th class="w-1/2 py-2 pl-6 text-left font-normal">
						{#if range}
							<span class="flex min-w-56 justify-between gap-2">
								<span>low {fmt(range.low)}</span><span>median {fmt(range.median)}</span><span
									>high {fmt(range.high)}</span
								>
							</span>
						{/if}
					</th>
				</tr>
			</thead>
			<tbody>
				{#each shown as p (p.symbol)}
					<tr class="border-b border-border/50 {p.symbol === symbol ? 'font-medium' : ''}">
						<td class="py-1.5" title={p.name}>
							<a href="/company/{p.symbol}" class="hover:underline">{p.symbol}</a>
							{#if p.symbol === symbol}<span class="text-xs text-muted-foreground">(this)</span
								>{/if}
						</td>
						<td class="py-1.5 text-right">{fmt(p.pe)}</td>
						<td class="py-1.5 text-right">{fmt(p.pb)}</td>
						<td class="py-1.5 pl-6">
							{#if range && p.pe !== undefined}
								<div class="relative h-1.5 min-w-56 rounded-full bg-muted">
									<div
										class="absolute top-0 h-1.5 w-px bg-muted-foreground"
										style="left: {at(range, range.median)}%"
									></div>
									<div
										class="absolute -top-0.5 size-2.5 -translate-x-1/2 rounded-full {p.symbol ===
										symbol
											? 'bg-foreground'
											: 'bg-muted-foreground'}"
										style="left: {at(range, p.pe)}%"
									></div>
								</div>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
	<p class="mt-2 text-xs text-muted-foreground">
		{#if windowed}
			<button
				onclick={() => (showAll = !showAll)}
				aria-expanded={showAll}
				class="underline underline-offset-2 hover:text-foreground"
				>{showAll
					? `Show the nearest ${WINDOW} each way`
					: `Show all ${peers.length} peers`}</button
			>
			·
		{/if}
		PE and PB from each company's latest filed quarter and latest close; this company's is {period ||
			'not recorded'}.
	</p>
</div>
