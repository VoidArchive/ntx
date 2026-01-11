<script lang="ts">
	import type { CorporateAction } from '$lib/gen/ntx/v1/common_pb';

	interface Props {
		actions?: CorporateAction[];
		class?: string;
	}

	let { actions = [], class: className = '' }: Props = $props();

	function fmt(value: number | undefined): string {
		if (value === undefined || value === 0) return '—';
		return value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	function fmtPercent(value: number | undefined): string {
		if (value === undefined || value === 0) return '—';
		return `${value.toFixed(1)}%`;
	}

	function fmtCash(value: number | undefined): string {
		if (value === undefined || value === 0) return '—';
		return `${value.toFixed(1)}%`;
	}

	// Sort by submitted date descending
	let sortedActions = $derived(
		[...actions].sort((a, b) => b.submittedDate.localeCompare(a.submittedDate)).slice(0, 10)
	);
</script>

{#if sortedActions.length > 0}
	<div class={className}>
		<h3 class="mb-4 font-serif text-base font-medium">Corporate Actions</h3>

		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-border text-left text-xs text-muted-foreground">
						<th class="pb-2 font-medium">Fiscal Year</th>
						<th class="pb-2 text-right font-medium">Bonus</th>
						<th class="pb-2 text-right font-medium">Rights</th>
						<th class="pb-2 text-right font-medium">Cash Dividend</th>
					</tr>
				</thead>
				<tbody>
					{#each sortedActions as a (a.id)}
						<tr class="border-b border-dotted border-border last:border-0">
							<td class="py-2.5 font-medium">{a.fiscalYear}</td>
							<td class="py-2.5 text-right text-positive tabular-nums">
								{fmtPercent(a.bonusPercentage)}
							</td>
							<td class="py-2.5 text-right tabular-nums">{fmtPercent(a.rightPercentage)}</td>
							<td class="py-2.5 text-right tabular-nums">{fmtCash(a.cashDividend)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
{:else}
	<div class={className}>
		<h3 class="mb-4 font-serif text-base font-medium">Corporate Actions</h3>
		<p class="text-sm text-muted-foreground">No corporate actions data available</p>
	</div>
{/if}
