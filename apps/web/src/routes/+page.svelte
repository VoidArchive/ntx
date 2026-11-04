<script lang="ts">
	import { formatLarge } from '#lib/utils.js';
	import type { Price } from '#lib/gen/ntx/v1/common_pb.js';

	let { data } = $props();

	function getPrice(companyId: bigint): Price | undefined {
		return data.prices?.find((p) => p.companyId === companyId);
	}

	// Sign and glyph carry direction, so colour is never the only signal.
	function signed(value: number | undefined, digits = 2): string {
		if (value === undefined) return '—';
		if (value > 0) return `▲ ${value.toFixed(digits)}%`;
		if (value < 0) return `▼ ${Math.abs(value).toFixed(digits)}%`;
		return `■ ${value.toFixed(digits)}%`;
	}

	function tone(value: number | undefined): string {
		if (value === undefined || value === 0) return 'text-muted-foreground';
		return value > 0 ? 'text-positive' : 'text-negative';
	}

	function ratio(value: number | undefined): string {
		return value === undefined ? '—' : `${value.toFixed(1)}x`;
	}

	let marketDay = $derived(data.marketDay);
	let sectors = $derived(data.overview?.sectors ?? []);
	let filedRecently = $derived(data.overview?.filedRecently ?? []);

	let largestMarketCap = $derived.by(() => {
		if (!data.companies) return [];
		return [...data.companies]
			.map((c) => {
				const price = getPrice(c.id);
				const shares = c.listedShares ? Number(c.listedShares) : 0;
				const ltp = price?.ltp ?? price?.close ?? 0;
				return {
					symbol: c.symbol,
					name: c.name,
					value: formatLarge(shares * ltp),
					cap: shares * ltp
				};
			})
			.filter((x) => x.cap > 0)
			.sort((a, b) => b.cap - a.cap)
			.slice(0, 5);
	});

	let mostTraded = $derived.by(() => {
		if (!data.prices) return [];
		return [...data.prices]
			.filter((p) => p.volume !== undefined && p.volume > 0)
			.sort((a, b) => Number(b.volume ?? 0) - Number(a.volume ?? 0))
			.slice(0, 5)
			.flatMap((p) => {
				const c = data.companies?.find((x) => x.id === p.companyId);
				if (!c) return [];
				return [{ symbol: c.symbol, name: c.name, value: Number(p.volume).toLocaleString() }];
			});
	});

	let movers = $derived([
		{
			title: 'Gainers',
			href: undefined,
			rows: (data.overview?.gainers ?? []).map((m) => ({
				symbol: m.symbol,
				name: m.name,
				value: signed(m.changePercent),
				tone: tone(m.changePercent)
			}))
		},
		{
			title: 'Losers',
			href: undefined,
			rows: (data.overview?.losers ?? []).map((m) => ({
				symbol: m.symbol,
				name: m.name,
				value: signed(m.changePercent),
				tone: tone(m.changePercent)
			}))
		},
		{
			title: 'Most traded',
			href: '/most-traded',
			rows: mostTraded.map((r) => ({ ...r, tone: 'text-muted-foreground' }))
		},
		{
			title: 'Largest market cap',
			href: '/market-cap',
			rows: largestMarketCap.map((r) => ({ ...r, tone: 'text-foreground' }))
		}
	]);
</script>

<div class="min-h-screen">
	<main class="mx-auto max-w-7xl px-4 py-8">
		{#if marketDay}
			<p
				class="mb-6 flex flex-wrap gap-x-4 gap-y-1 border-b border-border pb-3 text-sm text-muted-foreground"
			>
				<span class="text-foreground">{marketDay.businessDateBs} · {marketDay.businessDate}</span>
				<span>
					NEPSE <span class="text-foreground tabular-nums">{marketDay.indexClose.toFixed(2)}</span>
					<span
						class="tabular-nums {marketDay.indexChange > 0
							? 'text-positive'
							: marketDay.indexChange < 0
								? 'text-negative'
								: ''}"
					>
						{marketDay.indexChange > 0 ? '▲' : marketDay.indexChange < 0 ? '▼' : '■'}
						{Math.abs(marketDay.indexChange).toFixed(2)} ({Math.abs(
							marketDay.indexChangePercent
						).toFixed(2)}%)
					</span>
				</span>
				<span class="tabular-nums">
					<span class="text-positive">{marketDay.advances} up</span> ·
					<span class="text-negative">{marketDay.declines} down</span> ·
					{marketDay.unchanged} unchanged
				</span>
				<span
					>Turnover <span class="text-foreground tabular-nums"
						>Rs {formatLarge(marketDay.turnover)}</span
					></span
				>
			</p>
		{/if}

		{#if sectors.length > 0}
			<section>
				<h2 class="mb-3 font-serif text-base font-medium">Sectors</h2>
				<div class="overflow-x-auto">
					<table class="w-full text-sm tabular-nums">
						<thead class="text-left text-xs text-muted-foreground">
							<tr class="border-b border-border">
								<th class="py-2 pr-4 font-normal">Sector</th>
								<th class="py-2 pr-4 text-right font-normal">Companies</th>
								<th class="py-2 pr-4 text-right font-normal">Median PE</th>
								<th class="py-2 pr-4 text-right font-normal">Median PB</th>
								<th class="py-2 pr-4 text-right font-normal">Profitable</th>
								<th class="py-2 pr-4 text-right font-normal">Profit YoY</th>
								<th class="py-2 text-right font-normal">Today</th>
							</tr>
						</thead>
						<tbody>
							{#each sectors as s (s.name)}
								<tr class="border-b border-border/50 hover:bg-muted">
									<td class="py-2 pr-4">
										<a href="/company?sector={s.sector}" class="hover:underline">{s.name}</a>
									</td>
									<td class="py-2 pr-4 text-right">{s.companyCount}</td>
									<td class="py-2 pr-4 text-right">
										{ratio(s.medianPe)}
										<span class="text-xs text-muted-foreground">of {s.peCount} profitable</span>
									</td>
									<td class="py-2 pr-4 text-right">{ratio(s.medianPb)}</td>
									<td class="py-2 pr-4 text-right">
										{s.profitableShare === undefined
											? '—'
											: `${Math.round(s.profitableShare * 100)}%`}
										<span class="text-xs text-muted-foreground">of {s.quarterCount}</span>
									</td>
									<td class="py-2 pr-4 text-right {tone(s.profitChangeYoy)}">
										{signed(s.profitChangeYoy, 1)}
									</td>
									<td class="py-2 text-right {tone(s.dayChange)}">{signed(s.dayChange)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
				<p class="mt-2 text-xs text-muted-foreground">
					Medians over each company's latest quarter and latest close. Profit YoY sums year-to-date
					profit against the same quarter a year earlier.
				</p>
			</section>
		{/if}

		{#if filedRecently.length > 0}
			<section class="mt-12 border-t border-border pt-8">
				<h2 class="mb-3 font-serif text-base font-medium">Filed recently</h2>
				<ul class="divide-y divide-border/50 text-sm">
					{#each filedRecently as f (f.symbol)}
						<li class="flex flex-wrap items-baseline gap-x-3 py-2">
							<a href="/company/{f.symbol}" class="font-serif hover:underline">{f.symbol}</a>
							<span class="flex-1">{f.claim || f.name}</span>
							<span class="text-xs text-muted-foreground tabular-nums">
								{f.published ? 'published' : 'seen by NTX on'}
								{f.date}
							</span>
						</li>
					{/each}
				</ul>
			</section>
		{/if}

		<section class="mt-12 grid gap-6 border-t border-border pt-8 sm:grid-cols-2 lg:grid-cols-4">
			{#each movers as list (list.title)}
				<div>
					<div class="mb-2 flex items-baseline justify-between">
						<h3 class="font-serif text-sm font-medium">{list.title}</h3>
						{#if list.href}
							<a href={list.href} class="text-xs text-muted-foreground hover:underline">View all</a>
						{/if}
					</div>
					{#if list.rows.length === 0}
						<p class="text-xs text-muted-foreground">Nothing recorded for the latest session.</p>
					{/if}
					<ul class="text-sm">
						{#each list.rows as r (r.symbol)}
							<li>
								<a
									href="/company/{r.symbol}"
									class="flex items-baseline justify-between gap-2 py-1 hover:bg-muted"
									title={r.name}
								>
									<span class="font-serif">{r.symbol}</span>
									<span class="tabular-nums {r.tone}">{r.value}</span>
								</a>
							</li>
						{/each}
					</ul>
				</div>
			{/each}
			{#if marketDay}
				<p class="text-xs text-muted-foreground sm:col-span-2 lg:col-span-4">
					Session of {marketDay.businessDateBs} · {marketDay.businessDate}; market cap is listed
					shares times that close.
				</p>
			{/if}
		</section>
	</main>
</div>
