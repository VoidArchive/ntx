<script lang="ts">
	import type { Fund } from '#lib/types/fund.js';
	import FundStats from './FundStats.svelte';
	import { fundOpeningSentence, fundProvenance } from './story.js';

	interface Props {
		fund: Fund;
	}

	let { fund }: Props = $props();

	// NEPSE open-end funds issue units at Rs 10.
	let fromPar = $derived(((fund.nav_per_unit - 10) / 10) * 100);
	let initials = $derived(
		fund.fund_name
			.split(/\s+/)
			.filter((w) => /^[A-Za-z]/.test(w))
			.slice(0, 2)
			.map((w) => w[0])
			.join('')
			.toUpperCase()
	);
	function tone(value: number): string {
		return value > 0 ? 'text-positive' : value < 0 ? 'text-negative' : 'text-primary';
	}
</script>

<header class="card-report">
	<div class="grid gap-8 lg:grid-cols-[1fr_minmax(0,420px)]">
		<div class="min-w-0">
			<p class="text-sm text-muted-foreground">
				<a href="/mutual-funds" class="hover:text-foreground">Mutual funds</a>
				<span aria-hidden="true" class="mx-1">/</span>
				{fund.fund_manager}
			</p>
			<div class="mt-4 flex items-start gap-5">
				<div
					class="flex size-20 shrink-0 items-center justify-center rounded-lg bg-primary/15 font-serif text-3xl font-medium text-primary"
					aria-hidden="true"
				>
					{initials}
				</div>
				<div class="min-w-0">
					<h1 class="font-serif text-3xl font-medium tracking-tight sm:text-4xl">
						{fund.fund_name}
					</h1>
					<p class="mt-1 text-base text-muted-foreground">{fund.symbol} Fund Report</p>
					<p class="mt-1 text-sm text-muted-foreground">
						Open-end fund managed by {fund.fund_manager} · NAV report for {fund.report_date_nepali}
					</p>
				</div>
			</div>
			<p class="mt-6 max-w-3xl font-serif text-xl leading-snug">{fundOpeningSentence(fund)}</p>
			<p class="mt-6 text-xs text-muted-foreground">{fundProvenance(fund)}</p>
		</div>

		<div class="min-w-0 lg:pt-1">
			<p class="flex items-center gap-2 text-sm text-muted-foreground">
				<span class="size-2 rounded-full bg-primary" aria-hidden="true"></span> NAV per unit
			</p>
			<p class="mt-1 flex flex-wrap items-baseline gap-x-4">
				<span class="font-serif text-3xl font-medium tabular-nums"
					>Rs {fund.nav_per_unit.toFixed(2)}</span
				>
				<span class="text-sm tabular-nums">
					<span class="text-muted-foreground">from par</span>
					<span class={tone(fromPar)}>{fromPar > 0 ? '+' : ''}{fromPar.toFixed(2)}%</span>
				</span>
			</p>
			<div class="mt-5">
				<FundStats {fund} />
			</div>
		</div>
	</div>
</header>
