<script lang="ts">
	import type { Range, Strip } from '#lib/gen/ntx/v1/story_pb.js';
	import { at } from '#lib/utils.js';

	interface Props {
		strips: Strip[];
		// Newest filed period, which the company's own figures come from.
		period: string;
	}

	let { strips, period }: Props = $props();

	const LABELS: Record<string, string> = {
		pe: 'PE',
		pb: 'PB',
		eps: 'EPS',
		yield: 'Dividend yield'
	};

	// The one-word answer the bar draws.
	function reading(strip: Strip, range: Range | undefined): string {
		if (!range || strip.value === undefined) return '';
		const d = strip.value - range.median;
		if (Math.abs(d) < range.median * 0.02) return 'at the median';
		return d < 0 ? 'below the median' : 'above the median';
	}

	function fmt(ratio: string, value: number): string {
		const n = value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
		if (ratio === 'eps') return `Rs ${n}`;
		if (ratio === 'yield') return `${n}%`;
		return `${n}x`;
	}
</script>

{#snippet bar(strip: Strip, range: Range | undefined, label: string, missing: string)}
	<div class="flex justify-between gap-2 text-xs text-muted-foreground">
		<span>{label}</span>
		{#if range && strip.value !== undefined}
			<span class="text-foreground">{reading(strip, range)}</span>
		{/if}
	</div>
	{#if range}
		<div class="relative mt-1 h-2 rounded-full bg-muted">
			<div
				class="absolute -top-1 h-4 w-px bg-foreground/60"
				style="left: {at(range, range.median)}%"
				aria-hidden="true"
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
			<span>low {fmt(strip.ratio, range.low)}</span>
			<span>median {fmt(strip.ratio, range.median)}</span>
			<span>high {fmt(strip.ratio, range.high)}</span>
		</div>
	{:else}
		<div class="mt-1 text-xs text-muted-foreground">{missing}</div>
	{/if}
{/snippet}

<section id="strips" class="mt-6">
	<!-- How to read a strip, said once; each bar then only says where the dot landed. -->
	<p class="mb-4 text-xs text-muted-foreground">
		<span class="mr-1 inline-block size-2.5 rounded-full bg-foreground align-middle"></span> this company,
		on a bar from the group's lowest to its highest with the median at the centre. Left of centre is below
		the median; for PE and PB that is cheaper.
	</p>
	<div class="grid gap-x-10 gap-y-6 sm:grid-cols-2">
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
							? `Sector (${strip.sector.count} profitable)`
							: `Sector (${strip.sector?.count ?? 0})`,
						'No sector figures recorded'
					)}
					{@render bar(strip, strip.own, 'Own three years', 'No history recorded')}
				</div>
			</div>
		{/each}
		<p class="text-xs text-muted-foreground sm:col-span-2">
			This company's figures are from {period || 'no filing'} and the latest close. Sector ranges are
			over each sector company's latest filed quarter.
		</p>
	</div>
</section>
