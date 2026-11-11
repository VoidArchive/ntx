<script lang="ts">
	import { scaleLinear, scalePoint } from 'd3-scale';
	import type { Goal } from '#lib/gen/ntx/v1/portfolio_pb.js';

	// Value by month since the first transaction, then the band between the flat and
	// the with-return projections to the goal line. One encoding (DESIGN.md): the
	// latest value is the one orange point, the past and the band are Fog, the goal a
	// dashed Fog line. Inline SVG, as the snowflake is.
	let { goal, class: className = '' }: { goal: Goal; class?: string } = $props();

	// The width follows the container so axis text keeps its pixel size on a phone.
	let W = $state(640);
	const H = 220;
	const pad = { top: 12, right: 12, bottom: 28, left: 56 };

	let months = $derived([
		...goal.series.map((p) => p.month),
		...goal.band.slice(1).map((p) => p.month)
	]);
	let x = $derived(
		scalePoint<string>()
			.domain(months)
			.range([pad.left, W - pad.right])
	);
	let top = $derived(
		Math.max(
			goal.goalRs,
			...goal.series.map((p) => p.valueRs),
			...goal.band.map((p) => Math.max(p.flatRs, p.withReturnRs))
		) * 1.05
	);
	let y = $derived(
		scaleLinear()
			.domain([0, top])
			.range([H - pad.bottom, pad.top])
	);

	const line = (pts: { month: string; v: number }[]) =>
		pts.map((p, i) => `${i ? 'L' : 'M'}${x(p.month)},${y(p.v)}`).join(' ');
	let valueLine = $derived(line(goal.series.map((p) => ({ month: p.month, v: p.valueRs }))));
	let paidLine = $derived(line(goal.series.map((p) => ({ month: p.month, v: p.contributedRs }))));
	// The band is the flat path forward and the with-return path back, closed.
	let band = $derived.by(() => {
		if (goal.band.length < 2) return '';
		const up = goal.band.map((p) => `${x(p.month)},${y(Math.max(p.flatRs, p.withReturnRs))}`);
		const down = [...goal.band]
			.reverse()
			.map((p) => `${x(p.month)},${y(Math.min(p.flatRs, p.withReturnRs))}`);
		return `M${[...up, ...down].join(' L')} Z`;
	});
	let last = $derived(goal.series.at(-1));
	// Year labels at each January, plus the first month when no January sits near it.
	let ticks = $derived(
		months.filter(
			(m, i) => m.endsWith('-01') || (i === 0 && !months.slice(0, 3).some((x) => x.endsWith('-01')))
		)
	);
	const lakh = (v: number) => `${(v / 1e5).toFixed(v >= 1e5 ? 1 : 2)}`;
	let yTicks = $derived(y.ticks(4));
</script>

<div class={className} bind:clientWidth={W}>
	<svg
		viewBox="0 0 {W} {H}"
		width={W}
		height={H}
		role="img"
		aria-label="Value by month and the projection to the goal"
	>
		{#each yTicks as t (t)}
			<line x1={pad.left} x2={W - pad.right} y1={y(t)} y2={y(t)} class="stroke-border/60" />
			<text
				x={pad.left - 6}
				y={y(t) + 3}
				text-anchor="end"
				class="fill-muted-foreground text-[10px]">{lakh(t)} L</text
			>
		{/each}
		<line
			x1={pad.left}
			x2={W - pad.right}
			y1={y(goal.goalRs)}
			y2={y(goal.goalRs)}
			class="stroke-foreground/60"
			stroke-dasharray="4 4"
		/>
		<text
			x={W - pad.right}
			y={y(goal.goalRs) - 4}
			text-anchor="end"
			class="fill-foreground text-[10px]">goal Rs {lakh(goal.goalRs)} lakh</text
		>
		{#if band}
			<path d={band} class="fill-foreground/20" />
			<path
				d={line(goal.band.map((p) => ({ month: p.month, v: p.flatRs })))}
				class="fill-none stroke-foreground/40"
				stroke-dasharray="2 3"
			/>
		{/if}
		{#if goal.series.length > 1}
			<path d={paidLine} class="fill-none stroke-muted-foreground/50" stroke-width="1" />
			<path d={valueLine} class="fill-none stroke-foreground/70" stroke-width="1.75" />
		{/if}
		{#if last}
			<circle cx={x(last.month)} cy={y(last.valueRs)} r="3.5" class="fill-primary" />
		{/if}
		{#each ticks as m (m)}
			<text x={x(m)} y={H - 8} text-anchor="middle" class="fill-muted-foreground text-[10px]"
				>{m.endsWith('-01') ? m.slice(0, 4) : m}</text
			>
		{/each}
	</svg>
</div>
