<script lang="ts">
	import type { Fund } from '$lib/types/fund';
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import Building2 from '@lucide/svelte/icons/building-2';
	import Calendar from '@lucide/svelte/icons/calendar';

	interface Props {
		fund: Fund;
	}

	let { fund }: Props = $props();

	let navStatus = $derived.by(() => {
		if (fund.nav_per_unit > 10) return 'positive';
		if (fund.nav_per_unit < 10) return 'negative';
		return 'neutral';
	});
</script>

<div class="border-b border-border pb-6">
	<!-- Back link -->
	<a
		href="/mutual-funds"
		class="mb-4 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
	>
		<ArrowLeft class="size-4" />
		<span>All Funds</span>
	</a>

	<div class="flex flex-col gap-6 md:flex-row md:items-end md:justify-between">
		<!-- Left: Fund info -->
		<div>
			<div class="flex items-center gap-3">
				<h1 class="font-serif text-4xl tracking-tight md:text-5xl">{fund.symbol}</h1>
				<span
					class="rounded-full border px-3 py-1 text-xs font-medium {navStatus === 'positive'
						? 'border-positive/30 bg-positive/10 text-positive'
						: navStatus === 'negative'
							? 'border-negative/30 bg-negative/10 text-negative'
							: 'border-border bg-muted text-muted-foreground'}"
				>
					{navStatus === 'positive'
						? 'Above Par'
						: navStatus === 'negative'
							? 'Below Par'
							: 'At Par'}
				</span>
			</div>
			<p class="mt-2 text-lg text-muted-foreground">{fund.fund_name}</p>

			<!-- Meta -->
			<div class="mt-4 flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
				<span class="flex items-center gap-1.5">
					<Building2 class="size-4" />
					{fund.fund_manager}
				</span>
				<span class="flex items-center gap-1.5">
					<Calendar class="size-4" />
					{fund.report_date_english}
				</span>
			</div>
		</div>

		<!-- Right: NAV -->
		<div class="text-right">
			<p class="text-sm text-muted-foreground">NAV per Unit</p>
			<p class="font-serif text-5xl tabular-nums md:text-6xl">
				{fund.nav_per_unit.toFixed(2)}
			</p>
			{#if navStatus !== 'neutral'}
				<p
					class="mt-1 text-sm font-medium tabular-nums {navStatus === 'positive'
						? 'text-positive'
						: 'text-negative'}"
				>
					{navStatus === 'positive' ? '+' : ''}{(((fund.nav_per_unit - 10) / 10) * 100).toFixed(2)}%
					from par
				</p>
			{/if}
		</div>
	</div>
</div>
