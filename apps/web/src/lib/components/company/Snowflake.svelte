<script lang="ts">
	import type { Check } from '#lib/gen/ntx/v1/story_pb.js';
	import { AXES, checksFor, score } from './snowflake.js';

	interface Props {
		checks: Check[];
		class?: string;
	}

	let { checks, class: className = '' }: Props = $props();

	const C = 170; // centre
	const R = 118; // outer ring
	const CORE = 0.14; // an axis with no holds still shows a core, like a sprout

	let axes = $derived(
		AXES.map((axis, i) => {
			const s = score(checksFor(checks, axis.id));
			const angle = -Math.PI / 2 + (i * 2 * Math.PI) / AXES.length;
			const r = R * (CORE + (1 - CORE) * s.value);
			return { ...axis, ...s, angle, x: C + r * Math.cos(angle), y: C + r * Math.sin(angle) };
		})
	);

	// Closed Catmull-Rom spline through the axis points, so the shape is a blob, not a polygon.
	let blob = $derived.by(() => {
		const p = axes.map((a) => [a.x, a.y] as const);
		const n = p.length;
		let d = `M ${p[0][0]} ${p[0][1]}`;
		for (let i = 0; i < n; i++) {
			const p0 = p[(i - 1 + n) % n];
			const p1 = p[i];
			const p2 = p[(i + 1) % n];
			const p3 = p[(i + 2) % n];
			const c1 = [p1[0] + (p2[0] - p0[0]) / 6, p1[1] + (p2[1] - p0[1]) / 6];
			const c2 = [p2[0] - (p3[0] - p1[0]) / 6, p2[1] - (p3[1] - p1[1]) / 6];
			d += ` C ${c1[0]} ${c1[1]} ${c2[0]} ${c2[1]} ${p2[0]} ${p2[1]}`;
		}
		return d + ' Z';
	});

	// Each label rides an arc outside the ring; arcs in the lower half run the
	// other way so their text still reads left to right, upright.
	function labelArc(angle: number): string {
		const lower = Math.sin(angle) > 0.3;
		const r = lower ? R + 34 : R + 22;
		const spread = 0.6;
		const [a0, a1] = lower ? [angle + spread, angle - spread] : [angle - spread, angle + spread];
		const x0 = C + r * Math.cos(a0);
		const y0 = C + r * Math.sin(a0);
		const x1 = C + r * Math.cos(a1);
		const y1 = C + r * Math.sin(a1);
		return `M ${x0} ${y0} A ${r} ${r} 0 0 ${lower ? 0 : 1} ${x1} ${y1}`;
	}

	// "2 of 3 checks hold, 1 without history" names what the axis could not see.
	function describe(a: (typeof axes)[number]): string {
		const missing = a.total - a.evaluated;
		const base = `${a.holds} of ${a.total} checks hold`;
		return missing ? `${base}, ${missing} without history` : base;
	}

	let summary = $derived(axes.map((a) => `${a.label} ${describe(a)}`).join(', '));
	const uid = $props.id();
</script>

<svg
	viewBox="0 0 {C * 2} {C * 2}"
	class="h-auto w-full max-w-[340px] {className}"
	role="img"
	aria-label="Snowflake: {summary}"
>
	<defs>
		{#each axes as a (a.id)}
			<path id="{uid}-{a.id}" d={labelArc(a.angle)} fill="none" />
		{/each}
	</defs>

	<!-- Rings and spokes -->
	<circle cx={C} cy={C} r={R + 10} class="fill-none stroke-muted" stroke-width="12" opacity="0.5" />
	{#each [0.25, 0.5, 0.75, 1] as f (f)}
		<circle
			cx={C}
			cy={C}
			r={R * f}
			class="fill-none stroke-border"
			stroke-width={f === 1 ? 1.5 : 1}
		/>
	{/each}
	{#each axes as a (a.id)}
		<line
			x1={C}
			y1={C}
			x2={C + R * Math.cos(a.angle)}
			y2={C + R * Math.sin(a.angle)}
			class="stroke-border"
			stroke-width="1"
		/>
	{/each}

	<!-- The shape -->
	<path
		d={blob}
		class="fill-primary stroke-primary"
		fill-opacity="0.78"
		stroke-width="2"
		stroke-linejoin="round"
	/>
	<!-- An axis with a check that lacks history gets a hollow, dashed marker. -->
	{#each axes as a (a.id)}
		<circle
			cx={a.x}
			cy={a.y}
			r={a.evaluated < a.total ? 4 : 3}
			class="stroke-primary {a.evaluated < a.total ? 'fill-background' : 'fill-primary-foreground'}"
			stroke-width="1.5"
			stroke-dasharray={a.evaluated < a.total ? '2 1.5' : undefined}
		>
			<title>{a.label}: {describe(a)}</title>
		</circle>
	{/each}

	<!-- Labels -->
	{#each axes as a (a.id)}
		<text
			class="text-[13px] font-semibold tracking-[0.18em] uppercase {a.evaluated
				? 'fill-foreground'
				: 'fill-muted-foreground'}"
		>
			<textPath href="#{uid}-{a.id}" startOffset="50%" text-anchor="middle">{a.label}</textPath>
		</text>
	{/each}
</svg>
