<script lang="ts">
	import type { FundHolding } from '../../../routes/company/[symbol]/+page.server';

	interface Props {
		holdings: FundHolding[];
	}

	let { holdings }: Props = $props();

	// Get max value for scaling bars
	let maxValue = $derived(Math.max(...holdings.map((h) => h.value)));

	function fmtValue(value: number): string {
		if (value >= 10_000_000) return `${(value / 10_000_000).toFixed(2)} Cr`;
		if (value >= 100_000) return `${(value / 100_000).toFixed(2)} L`;
		return value.toLocaleString('en-NP');
	}
</script>

{#if holdings.length > 0}
	<div>
		<div class="mb-6 flex items-baseline justify-between">
			<h3 class="font-serif text-base font-medium">Mutual Fund Holdings</h3>
			<span class="text-sm text-muted-foreground">
				Held by {holdings.length} fund{holdings.length === 1 ? '' : 's'}
			</span>
		</div>

		<!-- Horizontal Bar Chart -->
		<div class="space-y-3">
			{#each holdings as holding (holding.fundSymbol)}
				{@const barWidth = (holding.value / maxValue) * 100}
				<a
					href="/mutual-funds/{holding.fundSymbol}"
					class="group flex items-center gap-3"
				>
					<!-- Fund Symbol -->
					<span class="w-16 shrink-0 text-sm font-medium transition-colors group-hover:text-primary">
						{holding.fundSymbol}
					</span>

					<!-- Bar -->
					<div class="relative h-6 flex-1">
						<div
							class="absolute inset-y-0 left-0 rounded bg-primary/80 transition-all group-hover:bg-primary"
							style="width: {barWidth}%"
						></div>
					</div>

					<!-- Value -->
					<span class="w-20 shrink-0 text-right text-sm tabular-nums text-muted-foreground">
						{fmtValue(holding.value)}
					</span>
				</a>
			{/each}
		</div>
	</div>
{/if}
