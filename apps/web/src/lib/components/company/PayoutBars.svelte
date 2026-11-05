<script lang="ts">
	import type { FiscalYear } from '#lib/gen/ntx/v1/story_pb.js';

	interface Props {
		years: FiscalYear[]; // newest first, as the API returns them
		class?: string;
	}

	let { years, class: className = '' }: Props = $props();

	let shown = $derived(years.slice(0, 8).reverse());
	let max = $derived(Math.max(...shown.map((y) => (y.cash ?? 0) + (y.bonus ?? 0)), 1));

	function pct(value: number | undefined): string {
		return value === undefined
			? '—'
			: `${value.toLocaleString('en-NP', { maximumFractionDigits: 2 })}%`;
	}
</script>

<!-- Labels carry the values; the bar only shows cash against bonus. A year whose
     notice never arrived is a gap, not a zero (GLOSSARY: Payout). -->
<div class={className}>
	<div
		class="flex h-40 items-end gap-2"
		role="img"
		aria-label="Cash and bonus dividend per fiscal year"
	>
		{#each shown as y (y.year)}
			{@const total = (y.cash ?? 0) + (y.bonus ?? 0)}
			<div class="flex min-w-0 flex-1 flex-col items-center justify-end text-center">
				<span class="mb-1 text-xs text-muted-foreground tabular-nums">
					{#if y.payout === 'not recorded'}—{:else}{pct(total)}{/if}
				</span>
				{#if y.payout === 'not recorded'}
					<div class="h-px w-full border-t border-dashed border-border"></div>
				{:else}
					<div
						class="flex w-full flex-col-reverse overflow-hidden rounded-t"
						style="height: {(total / max) * 100}%"
						title="Cash {pct(y.cash)} · Bonus {pct(y.bonus)}"
					>
						<div
							class="bg-foreground"
							style="height: {total ? ((y.cash ?? 0) / total) * 100 : 0}%"
						></div>
						<div
							class="bg-muted-foreground/40"
							style="height: {total ? ((y.bonus ?? 0) / total) * 100 : 0}%"
						></div>
					</div>
				{/if}
				<span class="mt-1 truncate text-xs text-muted-foreground tabular-nums">{y.year}</span>
			</div>
		{/each}
	</div>
	<p class="mt-2 text-xs text-muted-foreground">
		<span class="inline-block size-2 bg-foreground"></span> cash
		<span class="ml-3 inline-block size-2 bg-muted-foreground/40"></span> bonus, % of par · dashed: not
		recorded
	</p>
</div>
