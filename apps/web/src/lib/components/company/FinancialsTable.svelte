<script lang="ts">
	import type { Fundamental } from '#lib/gen/ntx/v1/common_pb.js';

	interface Props {
		fundamentals?: Fundamental[];
		class?: string;
	}

	let { fundamentals = [], class: className = '' }: Props = $props();

	function fmt(value: number | undefined): string {
		if (value === undefined || value === 0) return '—';
		return value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	function fmtLarge(value: number | undefined): string {
		if (value === undefined || value === 0) return '—';
		if (value >= 10_000_000) return `${(value / 10_000_000).toFixed(1)} Cr`;
		if (value >= 100_000) return `${(value / 100_000).toFixed(1)} L`;
		return fmt(value);
	}

	const yearFirst = (period: string) => period.split(' ').reverse().join(' ');

	// Rows as filed: profit is year-to-date and EPS annualized, unlike the bars above.
	let rows = $derived(
		fundamentals
			.filter((f) => !!f.quarter)
			// "Q1 2083/84" sorts as "2083/84 Q1": year first, then quarter.
			.sort((a, b) => yearFirst(b.period).localeCompare(yearFirst(a.period)))
			.slice(0, 8)
	);
</script>

<div class={className}>
	<h4 class="mb-2 text-xs font-medium text-muted-foreground">
		Rows as filed: year-to-date profit, annualized EPS
	</h4>

	{#if rows.length > 0}
		<div class="overflow-x-auto">
			<table class="w-full text-sm whitespace-nowrap [&_td+td]:pl-4 [&_th+th]:pl-4">
				<thead>
					<tr class="border-b border-border text-left text-xs text-muted-foreground">
						<th class="pb-2 font-medium"> Period </th>
						<th class="pb-2 text-right font-medium">EPS</th>
						<th class="pb-2 text-right font-medium">P/E</th>
						<th class="pb-2 text-right font-medium">Book Value</th>
						<th class="pb-2 text-right font-medium">Net Profit</th>
						<th class="pb-2 text-right font-medium">Published</th>
						<th class="pb-2 text-right font-medium">Source</th>
					</tr>
				</thead>
				<tbody>
					{#each rows as f (f.id)}
						<tr class="border-b border-dotted border-border last:border-0">
							<td class="py-2.5 font-medium">
								{f.period}
							</td>
							<td class="py-2.5 text-right tabular-nums">{fmt(f.eps)}</td>
							<td class="py-2.5 text-right tabular-nums">{fmt(f.peRatio)}</td>
							<td class="py-2.5 text-right tabular-nums">{fmt(f.bookValue)}</td>
							<td class="py-2.5 text-right tabular-nums">{fmtLarge(f.profitAmount)}</td>
							<td
								class="py-2.5 text-right tabular-nums {f.publishedAt
									? ''
									: 'text-muted-foreground'}"
							>
								{f.publishedAt ?? 'not recorded'}
							</td>
							<td class="py-2.5 text-right">
								{#if f.sourceUrl}
									<a
										href={f.sourceUrl}
										target="_blank"
										rel="noopener noreferrer"
										class="underline underline-offset-2 hover:text-foreground">filing</a
									>
								{:else}
									<span class="text-muted-foreground">not recorded</span>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
