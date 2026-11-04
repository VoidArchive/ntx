<script lang="ts">
	import type { Quarter } from '#lib/gen/ntx/v1/story_pb.js';

	interface Props {
		quarters: Quarter[];
		class?: string;
	}

	let { quarters, class: className = '' }: Props = $props();

	// Eight quarters on wide screens; the CSS below hides the oldest four on phones.
	let shown = $derived(quarters.slice(-8));
	let maxProfit = $derived(Math.max(...shown.map((q) => Math.abs(q.profit)), 1));

	function crore(rs: number): string {
		return (rs / 1e7).toFixed(1);
	}
</script>

{#if shown.length > 0}
	<div class={className}>
		<div class="flex items-end gap-2">
			{#each shown as q, i (q.period)}
				<div
					class="min-w-0 flex-1 flex-col items-center justify-end text-center {i < shown.length - 4
						? 'hidden sm:flex'
						: 'flex'}"
					title="{q.period}: profit Rs {crore(q.profit)} Cr{q.isMonsoon ? ', monsoon quarter' : ''}"
				>
					<span class="mb-1 min-h-[3.75em] text-[10px] leading-tight text-chart-4">{q.label}</span>
					<span class="text-[11px] tabular-nums {q.profit < 0 ? 'text-negative' : ''}">
						{q.profit < 0 ? '▼ ' : ''}{crore(q.profit)}
					</span>
					<div class="flex h-40 w-full items-end justify-center">
						<div
							class="w-full max-w-9 rounded-t {q.profit < 0
								? 'bg-negative/70'
								: q.isMonsoon
									? 'bg-positive/40'
									: 'bg-positive/80'}"
							style="height: {Math.max((Math.abs(q.profit) / maxProfit) * 100, 1)}%"
						></div>
					</div>
					<span class="mt-1 text-[10px] text-muted-foreground">{q.period}</span>
					<span class="text-[10px] text-muted-foreground tabular-nums">
						{q.eps !== undefined ? `EPS ${q.eps.toFixed(2)}` : ''}
					</span>
				</div>
			{/each}
		</div>
		<p class="mt-3 text-xs text-muted-foreground">
			Quarter-only net profit in Rs crore, derived from year-to-date filings. EPS is for the
			quarter.{shown.some((q) => q.isMonsoon) ? ' Lighter bars are monsoon quarters.' : ''}
		</p>
	</div>
{:else}
	<div
		class="flex h-[280px] items-center justify-center rounded-xl border border-dashed border-border"
	>
		<p class="text-sm text-muted-foreground">No earnings data available</p>
	</div>
{/if}
