<script lang="ts">
	import type { Range, Strip } from '#lib/gen/ntx/v1/story_pb.js';

	interface Props {
		strips: Strip[];
	}

	let { strips }: Props = $props();

	const LABELS: Record<string, string> = {
		pe: 'PE',
		pb: 'PB',
		eps: 'EPS',
		yield: 'Dividend yield'
	};

	function fmt(ratio: string, value: number): string {
		const n = value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
		if (ratio === 'eps') return `Rs ${n}`;
		if (ratio === 'yield') return `${n}%`;
		return `${n}x`;
	}

	// Position on the bar as a percentage, clamped so outliers sit at the ends.
	function at(range: Range, value: number): number {
		if (range.high === range.low) return 50;
		return Math.min(100, Math.max(0, ((value - range.low) / (range.high - range.low)) * 100));
	}
</script>

{#snippet bar(strip: Strip, range: Range | undefined, label: string, missing: string)}
	<div class="text-xs text-muted-foreground">{label}</div>
	{#if range}
		<div class="relative mt-1 h-2 rounded-full bg-muted">
			<div
				class="absolute top-0 h-2 w-px bg-muted-foreground"
				style="left: {at(range, range.median)}%"
			></div>
			{#if strip.value !== undefined}
				<div
					class="absolute -top-1 size-4 -translate-x-1/2 rounded-full border-2 border-background bg-foreground"
					style="left: {at(range, strip.value)}%"
					title="This company: {fmt(strip.ratio, strip.value)}"
				></div>
			{/if}
		</div>
		<div class="mt-1 flex justify-between text-xs text-muted-foreground tabular-nums">
			<span>{fmt(strip.ratio, range.low)}</span>
			<span>median {fmt(strip.ratio, range.median)}</span>
			<span>{fmt(strip.ratio, range.high)}</span>
		</div>
	{:else}
		<div class="mt-1 text-xs text-muted-foreground">{missing}</div>
	{/if}
{/snippet}

<section class="mt-6 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
	{#each strips as strip (strip.ratio)}
		<div id="strip-{strip.ratio}" class="min-w-0">
			<div class="flex items-baseline justify-between">
				<h3 class="font-serif text-base font-medium">{LABELS[strip.ratio] ?? strip.ratio}</h3>
				<span class="tabular-nums"
					>{strip.value !== undefined ? fmt(strip.ratio, strip.value) : '—'}</span
				>
			</div>
			<div class="mt-2 space-y-3">
				{@render bar(
					strip,
					strip.sector,
					strip.ratio === 'pe' && strip.sector
						? `Sector, of ${strip.sector.count} profitable`
						: `Sector (${strip.sector?.count ?? 0})`,
					'No sector figures recorded'
				)}
				{@render bar(strip, strip.own, 'Own three years', 'No history recorded')}
			</div>
		</div>
	{/each}
</section>
