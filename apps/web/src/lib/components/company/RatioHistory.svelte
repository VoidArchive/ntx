<script lang="ts">
	import type { Fundamental } from '#lib/gen/ntx/v1/common_pb.js';

	interface Props {
		fundamentals: Fundamental[]; // quarterly rows, any order
		median?: number; // the company's own three-year median PE
		current?: number; // PE at the latest close, which moves daily while the filed figure does not
		class?: string;
	}

	let { fundamentals, median, current, class: className = '' }: Props = $props();

	// "Q1 2083/84" sorts as "2083/84 Q1": year first, then quarter.
	const yearFirst = (period: string) => period.split(' ').reverse().join(' ');

	// The last twelve quarters with a PE; a loss quarter has none and leaves a gap.
	let rows = $derived(
		fundamentals
			.filter((f) => !!f.quarter && f.peRatio !== undefined && f.peRatio > 0)
			.sort((a, b) => yearFirst(a.period).localeCompare(yearFirst(b.period)))
			.slice(-12)
	);
	let max = $derived(Math.max(...rows.map((r) => r.peRatio ?? 0), median ?? 0) * 1.15);
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
			PE as filed each quarter{#if current !== undefined}; at the latest close it is {fmt(
					current
				)}{/if}.
		</p>
	</div>
{:else}
	<p class="text-sm text-muted-foreground {className}">
		No quarter with a positive EPS, so no PE to show.
	</p>
{/if}
