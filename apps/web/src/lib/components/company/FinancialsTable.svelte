<script lang="ts">
	import type { Fundamental } from '$lib/gen/ntx/v1/common_pb';

	export type ViewMode = 'quarterly' | 'annually';

	interface Props {
		fundamentals?: Fundamental[];
		viewMode?: ViewMode;
		onViewModeChange?: (mode: ViewMode) => void;
		class?: string;
	}

	let {
		fundamentals = [],
		viewMode = 'quarterly',
		onViewModeChange,
		class: className = ''
	}: Props = $props();

	function setViewMode(mode: ViewMode) {
		onViewModeChange?.(mode);
	}

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

	function formatQuarter(quarter: string | undefined): string {
		if (!quarter) return '';
		const q = quarter.toLowerCase();
		if (q.includes('first') || q === '1' || q === 'q1') return 'Q1';
		if (q.includes('second') || q === '2' || q === 'q2') return 'Q2';
		if (q.includes('third') || q === '3' || q === 'q3') return 'Q3';
		if (q.includes('fourth') || q === '4' || q === 'q4') return 'Q4';
		return quarter;
	}

	function quarterToNumber(quarter: string | undefined): number {
		if (!quarter) return 5;
		const q = quarter.toLowerCase();
		if (q.includes('fourth') || q === '4' || q === 'q4') return 4;
		if (q.includes('third') || q === '3' || q === 'q3') return 3;
		if (q.includes('second') || q === '2' || q === 'q2') return 2;
		if (q.includes('first') || q === '1' || q === 'q1') return 1;
		return 0;
	}

	function isQuarterly(f: Fundamental): boolean {
		return !!f.quarter;
	}

	let filteredFundamentals = $derived(
		[...fundamentals]
			.filter((f) => (viewMode === 'quarterly' ? isQuarterly(f) : !isQuarterly(f)))
			.sort((a, b) => {
				const yearCompare = b.fiscalYear.localeCompare(a.fiscalYear);
				if (yearCompare !== 0) return yearCompare;
				return quarterToNumber(b.quarter) - quarterToNumber(a.quarter);
			})
			.slice(0, 5)
	);

	let hasQuarterly = $derived(fundamentals.some(isQuarterly));
	let hasAnnual = $derived(fundamentals.some((f) => !isQuarterly(f)));
</script>

<div class={className}>
	<div class="mb-4 flex items-center justify-between">
		<h3 class="font-serif text-base font-medium">Financial History</h3>

		{#if hasQuarterly || hasAnnual}
			<div class="flex rounded-md border border-border text-xs">
				<button
					class="px-3 py-1 transition-colors {viewMode === 'quarterly'
						? 'bg-foreground text-background'
						: 'hover:bg-muted'}"
					onclick={() => setViewMode('quarterly')}
					disabled={!hasQuarterly}
				>
					Quarterly
				</button>
				<button
					class="px-3 py-1 transition-colors {viewMode === 'annually'
						? 'bg-foreground text-background'
						: 'hover:bg-muted'}"
					onclick={() => setViewMode('annually')}
					disabled={!hasAnnual}
				>
					Annually
				</button>
			</div>
		{/if}
	</div>

	{#if filteredFundamentals.length > 0}
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-border text-left text-xs text-muted-foreground">
						<th class="pb-2 font-medium">
							{viewMode === 'quarterly' ? 'Period' : 'Fiscal Year'}
						</th>
						<th class="pb-2 text-right font-medium">EPS</th>
						<th class="pb-2 text-right font-medium">P/E</th>
						<th class="pb-2 text-right font-medium">Book Value</th>
						<th class="pb-2 text-right font-medium">Net Profit</th>
					</tr>
				</thead>
				<tbody>
					{#each filteredFundamentals as f (f.id)}
						<tr class="border-b border-dotted border-border last:border-0">
							<td class="py-2.5 font-medium">
								{#if viewMode === 'quarterly'}
									{formatQuarter(f.quarter)} {f.fiscalYear}
								{:else}
									{f.fiscalYear}
								{/if}
							</td>
							<td class="py-2.5 text-right tabular-nums">{fmt(f.eps)}</td>
							<td class="py-2.5 text-right tabular-nums">{fmt(f.peRatio)}</td>
							<td class="py-2.5 text-right tabular-nums">{fmt(f.bookValue)}</td>
							<td class="py-2.5 text-right tabular-nums">{fmtLarge(f.profitAmount)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{:else}
		<p class="text-sm text-muted-foreground">
			No {viewMode === 'quarterly' ? 'quarterly' : 'annual'} data available
		</p>
	{/if}
</div>
