<script lang="ts">
	import type { FiscalYear, Growth } from '#lib/gen/ntx/v1/story_pb.js';

	interface Props {
		years: FiscalYear[];
		growth: Growth[];
	}

	let { years, growth }: Props = $props();

	const METRICS: Record<string, string> = {
		profit: 'Net profit',
		eps: 'EPS',
		book_value: 'Book value',
		price: 'Price, bonus-adjusted'
	};

	function num(value: number | undefined, digits = 2): string {
		if (value === undefined) return '—';
		return value.toLocaleString('en-NP', {
			minimumFractionDigits: digits,
			maximumFractionDigits: digits
		});
	}

	function crore(value: number | undefined): string {
		return value === undefined ? '—' : num(value / 1e7);
	}

	function pct(value: number | undefined): string {
		if (value === undefined) return '—';
		return `${value > 0 ? '+' : ''}${num(value, 1)}%`;
	}

	// Colour is never the only carrier: the return also gets a glyph.
	function arrow(value: number | undefined): string {
		if (value === undefined || value === 0) return '';
		return value > 0 ? '▲ ' : '▼ ';
	}

	// A missing payout figure reads as the year's status, never as zero.
	function part(row: FiscalYear, value: number | undefined): string {
		if (value !== undefined) return `${num(value, 1)}%`;
		return row.payout === 'not recorded' ? 'n/r' : '0%';
	}
</script>

<section id="fiscal-years" class="grid gap-8 lg:grid-cols-12">
	<div class="min-w-0 overflow-x-auto lg:col-span-8">
		<h3 class="mb-4 font-serif text-base font-medium">Fiscal years</h3>
		{#if years.length === 0}
			<p class="text-sm text-muted-foreground">No full-year filings or notices recorded.</p>
		{:else}
			<table class="w-full text-sm whitespace-nowrap tabular-nums [&_td+td]:pl-3 [&_th+th]:pl-3">
				<thead class="text-xs text-muted-foreground">
					<tr class="border-b border-border">
						<th class="py-2 text-left font-normal">FY</th>
						<th class="py-2 text-right font-normal">EPS</th>
						<th class="py-2 text-right font-normal">Profit (Rs cr)</th>
						<th class="py-2 text-right font-normal">Book value</th>
						<th class="py-2 text-right font-normal">Bonus</th>
						<th class="py-2 text-right font-normal">Cash</th>
						<th class="py-2 text-right font-normal">Rights*</th>
						<th class="py-2 text-right font-normal">Shares (cr)</th>
						<th class="py-2 text-right font-normal">Return</th>
					</tr>
				</thead>
				<tbody>
					{#each years as row (row.year)}
						<tr class="border-b border-border/50">
							<td class="py-2">
								{#if row.source}
									<a href={row.source} class="underline decoration-dotted underline-offset-4"
										>{row.year}</a
									>
								{:else}{row.year}{/if}
							</td>
							<td class="py-2 text-right">{num(row.eps)}</td>
							<td class="py-2 text-right">{crore(row.profit)}</td>
							<td class="py-2 text-right">{num(row.bookValue)}</td>
							<td class="py-2 text-right">{part(row, row.bonus)}</td>
							<td class="py-2 text-right">{part(row, row.cash)}</td>
							<td class="py-2 text-right">{part(row, row.rights)}</td>
							<td class="py-2 text-right">{crore(row.sharesOutstanding)}</td>
							<td
								class="py-2 text-right {row.return === undefined
									? ''
									: row.return >= 0
										? 'text-positive'
										: 'text-negative'}">{arrow(row.return)}{pct(row.return)}</td
							>
						</tr>
					{/each}
				</tbody>
			</table>
			<p class="mt-2 text-xs text-muted-foreground">
				Return is close at fiscal-year end with bonus shares and cash added, over close at the
				start. *Rights are shown but excluded from the return. n/r: payout not recorded. Full-year
				figures come from the annual report, else the Q4 filing.
			</p>
		{/if}
	</div>

	<div class="min-w-0 overflow-x-auto lg:col-span-4">
		<h3 class="mb-4 font-serif text-base font-medium">Compounded growth, per year</h3>
		<table class="w-full text-sm whitespace-nowrap tabular-nums [&_td+td]:pl-3 [&_th+th]:pl-3">
			<thead class="text-xs text-muted-foreground">
				<tr class="border-b border-border">
					<th class="py-2 text-left font-normal"></th>
					<th class="py-2 text-right font-normal">1y</th>
					<th class="py-2 text-right font-normal">3y</th>
					<th class="py-2 text-right font-normal">5y</th>
				</tr>
			</thead>
			<tbody>
				{#each growth as g (g.metric)}
					<tr class="border-b border-border/50">
						<td class="py-2">{METRICS[g.metric] ?? g.metric}</td>
						<td class="py-2 text-right">{pct(g.one)}</td>
						<td class="py-2 text-right">{pct(g.three)}</td>
						<td class="py-2 text-right">{pct(g.five)}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		<p class="mt-2 text-xs text-muted-foreground">— where a year in the span is missing.</p>
	</div>
</section>
