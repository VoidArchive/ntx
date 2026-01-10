<script lang="ts">
	import type { Fundamental } from '$lib/gen/ntx/v1/common_pb';

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

	// Sort by fiscal year descending and take latest 5
	let sortedFundamentals = $derived(
		[...fundamentals]
			.sort((a, b) => b.fiscalYear.localeCompare(a.fiscalYear))
			.slice(0, 5)
	);
</script>

{#if sortedFundamentals.length > 0}
	<div class={className}>
		<h3 class="mb-4 font-serif text-base font-medium">Financial History</h3>

		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-border text-left text-xs text-muted-foreground">
						<th class="pb-2 font-medium">Fiscal Year</th>
						<th class="pb-2 text-right font-medium">EPS</th>
						<th class="pb-2 text-right font-medium">P/E</th>
						<th class="pb-2 text-right font-medium">Book Value</th>
						<th class="pb-2 text-right font-medium">Net Profit</th>
					</tr>
				</thead>
				<tbody>
					{#each sortedFundamentals as f (f.fiscalYear)}
						<tr class="border-b border-dotted border-border last:border-0">
							<td class="py-2.5 font-medium">{f.fiscalYear}{f.quarter ? ` Q${f.quarter}` : ''}</td>
							<td class="py-2.5 text-right tabular-nums">{fmt(f.eps)}</td>
							<td class="py-2.5 text-right tabular-nums">{fmt(f.peRatio)}</td>
							<td class="py-2.5 text-right tabular-nums">{fmt(f.bookValue)}</td>
							<td class="py-2.5 text-right tabular-nums">{fmtLarge(f.profitAmount)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
{:else}
	<div class={className}>
		<h3 class="mb-4 font-serif text-base font-medium">Financial History</h3>
		<p class="text-sm text-muted-foreground">No financial data available</p>
	</div>
{/if}
