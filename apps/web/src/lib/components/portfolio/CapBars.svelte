<script lang="ts">
	import type { Holding, Rules } from '#lib/gen/ntx/v1/portfolio_pb.js';

	// Each holding's weight against the soft and hard caps, then each sector against
	// the sector cap, as Fog bars with the caps as ticks. Red only past a cap the API
	// calls Triggered (DESIGN.md, the Direction Rule); between the caps is muted.
	let { holdings, rules }: { holdings: Holding[]; rules: Rules } = $props();

	const pct = (v: number) => `${v.toLocaleString('en-NP', { maximumFractionDigits: 1 })}%`;

	let priced = $derived(holdings.filter((h) => h.quantity > 0n && h.close !== undefined));
	let sectors = $derived.by(() => {
		const sum: Record<string, number> = {};
		for (const h of priced) sum[h.sector] = (sum[h.sector] ?? 0) + h.weightPercent;
		return Object.entries(sum).sort((a, b) => b[1] - a[1]);
	});
	// One scale for every bar: the larger of the hard cap, the sector cap and the top weight.
	let scale = $derived(
		Math.max(
			rules.hardCapPercent,
			rules.sectorCapPercent,
			...priced.map((h) => h.weightPercent),
			...sectors.map(([, w]) => w)
		) * 1.1
	);
	const x = (v: number) => `${Math.min(100, (v / scale) * 100)}%`;
	const fill = (weight: number, cap: number, watch?: number) =>
		weight > cap
			? 'bg-negative'
			: watch !== undefined && weight > watch
				? 'bg-muted-foreground'
				: 'bg-foreground/25';
</script>

<div class="grid gap-x-8 gap-y-1 text-sm lg:grid-cols-2">
	<div>
		<h3 class="text-xs text-muted-foreground">
			Holdings against the {pct(rules.softCapPercent)} soft and {pct(rules.hardCapPercent)} hard cap
		</h3>
		<ul class="mt-2 space-y-1">
			{#each priced.filter((h) => !h.isFund) as h (h.symbol)}
				<li class="grid grid-cols-[4.5rem_1fr_3.5rem] items-center gap-2">
					<span class="font-medium">{h.symbol}</span>
					<div class="relative h-3 rounded-sm bg-accent/40">
						<div
							class="h-3 rounded-sm {fill(
								h.weightPercent,
								rules.hardCapPercent,
								rules.softCapPercent
							)}"
							style="width: {x(h.weightPercent)}"
						></div>
						<span
							class="absolute top-0 h-3 border-l border-dashed border-muted-foreground"
							style="left: {x(rules.softCapPercent)}"
						></span>
						<span
							class="absolute top-0 h-3 border-l border-foreground"
							style="left: {x(rules.hardCapPercent)}"
						></span>
					</div>
					<span class="text-right tabular-nums">{pct(h.weightPercent)}</span>
				</li>
			{:else}
				<li class="text-muted-foreground">No priced holding.</li>
			{/each}
		</ul>
	</div>
	<div>
		<h3 class="text-xs text-muted-foreground">
			Sectors against the {pct(rules.sectorCapPercent)} sector cap
		</h3>
		<ul class="mt-2 space-y-1">
			{#each sectors as [sector, weight] (sector)}
				<li
					class="grid grid-cols-[1fr_1fr_3.5rem] items-center gap-2 sm:grid-cols-[9rem_1fr_3.5rem]"
				>
					<span class="truncate text-muted-foreground">{sector}</span>
					<div class="relative h-3 rounded-sm bg-accent/40">
						<div
							class="h-3 rounded-sm {fill(weight, rules.sectorCapPercent)}"
							style="width: {x(weight)}"
						></div>
						<span
							class="absolute top-0 h-3 border-l border-foreground"
							style="left: {x(rules.sectorCapPercent)}"
						></span>
					</div>
					<span class="text-right tabular-nums">{pct(weight)}</span>
				</li>
			{:else}
				<li class="text-muted-foreground">No sector held.</li>
			{/each}
		</ul>
	</div>
</div>
