<script lang="ts">
	import type { Quarter } from '#lib/gen/ntx/v1/story_pb.js';

	interface Props {
		quarters: Quarter[]; // oldest first, as the story returns them
		median?: number; // the company's own three-year median PE, on the same trailing basis
		current?: number; // PE at the latest close, which moves daily while the filed figure does not
		class?: string;
	}

	let { quarters, median, current, class: className = '' }: Props = $props();

	// The last twelve quarters priced on trailing EPS (GLOSSARY: Trailing EPS); a quarter
	// without four consecutive quarters behind it, or with a trailing loss, leaves a gap.
	let rows = $derived(
		quarters
			.filter((q) => q.trailingPe !== undefined && q.trailingPe > 0)
			.map((q) => ({ period: q.period, peRatio: q.trailingPe as number }))
			.slice(-12)
	);
	let max = $derived(Math.max(...rows.map((r) => r.peRatio), median ?? 0) * 1.15);
	const fmt = (v: number) => `${v.toLocaleString('en-NP', { maximumFractionDigits: 2 })}x`;
</script>

{#if rows.length > 0}
	<div class={className}>
		<div class="relative">
			{#if median !== undefined}
				<!-- The median as one dashed line; the bars answer the question against it. -->
				<div
					class="absolute inset-x-0 border-t border-dashed border-foreground/60"
					style="bottom: {(median / max) * 100}%"
					aria-hidden="true"
				>
					<span
						class="absolute -top-5 right-0 rounded bg-card px-1 text-[11px] text-muted-foreground tabular-nums"
						>median {fmt(median)}</span
					>
				</div>
			{/if}
			<div class="flex h-44 items-end gap-2">
				{#each rows as r, i (r.period)}
					{@const latest = i === rows.length - 1}
					<div
						class="h-full min-w-0 flex-1 flex-col items-center justify-end {i < rows.length - 6
							? 'hidden sm:flex'
							: 'flex'}"
						title="{r.period}: PE {fmt(r.peRatio ?? 0)}"
					>
						<span
							class="mb-1 text-[11px] tabular-nums {latest
								? 'font-medium'
								: 'text-muted-foreground'}">{fmt(r.peRatio ?? 0)}</span
						>
						<div
							class="w-full max-w-10 rounded-t-sm {latest ? 'bg-primary' : 'bg-foreground/25'}"
							style="height: {Math.max(((r.peRatio ?? 0) / max) * 100, 1)}%"
						></div>
					</div>
				{/each}
			</div>
		</div>
		<div class="mt-1 flex gap-2">
			{#each rows as r, i (r.period)}
				<span
					class="min-w-0 flex-1 text-center text-[10px] text-muted-foreground {i < rows.length - 6
						? 'hidden sm:block'
						: 'block'}">{r.period}</span
				>
			{/each}
		</div>
		<p class="mt-3 text-xs text-muted-foreground">
			PE at each filing's price on the four quarters to it{#if current !== undefined}; at the latest
				close it is {fmt(current)}{/if}.
		</p>
	</div>
{:else}
	<p class="text-sm text-muted-foreground {className}">
		No four consecutive quarters with a profit, so no trailing PE to show.
	</p>
{/if}
