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

	// The sentence first, as on every page: the session when one is recorded, else the market's shape.
	let listed = $derived(sectors.reduce((n, s) => n + s.companyCount, 0));
	let filed = $derived(sectors.reduce((n, s) => n + s.quarterCount, 0));
	let profitable = $derived(
		sectors.reduce((n, s) => n + Math.round((s.profitableShare ?? 0) * s.quarterCount), 0)
	);
	let breadth = $derived(
		marketDay ? marketDay.advances + marketDay.declines + marketDay.unchanged : 0
	);
	function word(value: number): string {
		return value > 0 ? 'up' : value < 0 ? 'down' : 'flat';
	}
	let filedRecently = $derived(data.overview?.filedRecently ?? []);

	// One date heading per run of filings, instead of the same date on every row.
	let filedGroups = $derived.by(() => {
		const groups: { label: string; rows: typeof filedRecently }[] = [];
		for (const f of filedRecently) {
			const label = `${f.published ? 'Published' : 'Seen by NTX on'} ${f.date}`;
			const last = groups.at(-1);
			if (last?.label === label) last.rows.push(f);
			else groups.push({ label, rows: [f] });
		}
		return groups;
	});

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
			href: '/company?sort=turnover',
			rows: mostTraded.map((r) => ({ ...r, tone: 'text-muted-foreground' }))
		},
		{
			title: 'Largest market cap',
			href: '/company?sort=market_cap',
			rows: largestMarketCap.map((r) => ({ ...r, tone: 'text-foreground' }))
		}
	]);
</script>

<div class="min-h-screen">
	<main class="mx-auto max-w-7xl px-4 py-8">
		<header class="border-b border-border pb-8">
			{#if marketDay}
				<h1 class="font-serif text-3xl font-medium tracking-tight">
					{marketDay.businessDateBs}
					<span class="text-lg font-normal text-muted-foreground">· {marketDay.businessDate}</span>
				</h1>
				<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">
					NEPSE closed at <span class="tabular-nums"
						>{marketDay.indexClose.toLocaleString('en-NP', {
							minimumFractionDigits: 2,
							maximumFractionDigits: 2
						})}</span
					>,
					<span
						class="tabular-nums {marketDay.indexChange > 0
							? 'text-positive'
							: marketDay.indexChange < 0
								? 'text-negative'
								: ''}"
						>{word(marketDay.indexChange)}
						{Math.abs(marketDay.indexChangePercent).toFixed(2)}%</span
					>, on turnover of Rs {formatLarge(marketDay.turnover)}. {profitable} of {filed} companies with
					a filed quarter are profitable.
				</p>
				<!-- Breadth: one bar, three segments, each named so colour is not the only carrier. -->
				{#if breadth > 0}
					<div
						class="mt-4 flex h-2 max-w-3xl overflow-hidden rounded-full bg-muted"
						role="img"
						aria-label="{marketDay.advances} advanced, {marketDay.declines} declined, {marketDay.unchanged} unchanged"
					>
						<div class="bg-positive" style="width: {(marketDay.advances / breadth) * 100}%"></div>
						<div
							class="border-l-2 border-background bg-negative"
							style="width: {(marketDay.declines / breadth) * 100}%"
						></div>
					</div>
					<p class="mt-1 flex max-w-3xl justify-between text-xs text-muted-foreground tabular-nums">
						<span><span class="text-positive">▲</span> {marketDay.advances} advanced</span>
						<span>{marketDay.unchanged} unchanged</span>
						<span>{marketDay.declines} declined <span class="text-negative">▼</span></span>
					</p>
				{/if}
			{:else}
				<h1 class="font-serif text-3xl font-medium tracking-tight">NEPSE</h1>
				<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">
					{listed} listed companies in {sectors.length} sectors; {profitable} of {filed} with a filed
					quarter are profitable.
				</p>
				<p class="mt-2 text-sm text-muted-foreground">No session summary recorded yet.</p>
			{/if}
		</header>

		{#if sectors.length > 0}
			<section class="mt-10">
				<h2 class="mb-3 font-serif text-base font-medium">Sectors</h2>
				<div class="overflow-x-auto">
					<table class="w-full text-sm whitespace-nowrap tabular-nums">
						<thead class="text-left text-xs text-muted-foreground">
							<tr class="border-b border-border">
								<th class="sticky left-0 bg-background py-2 pr-4 font-normal">Sector</th>
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
									<!-- Pinned so a phone can scroll the figures and keep the row's name. -->
									<td class="sticky left-0 bg-background py-2 pr-4">
										<a href="/company?sector={s.sector}" class="hover:underline">{s.name}</a>
									</td>
									<td class="py-2 pr-4 text-right">{s.companyCount}</td>
									<td class="py-2 pr-4 text-right">
										{ratio(s.medianPe)}
										<span class="text-xs text-muted-foreground">of {s.peCount} profitable</span>
									</td>
									<td class="py-2 pr-4 text-right">{ratio(s.medianPb)}</td>
									<td class="py-2 pr-4 text-right">
										<!-- The share as a bar beside the number, so the column reads without arithmetic. -->
										{#if s.profitableShare !== undefined}
											<span
												class="mr-2 inline-block h-1.5 w-16 overflow-hidden rounded-full bg-muted align-middle"
												aria-hidden="true"
											>
												<span
													class="block h-full bg-foreground/70"
													style="width: {s.profitableShare * 100}%"
												></span>
											</span>
										{/if}
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

		{#if filedGroups.length > 0}
			<section class="mt-12 border-t border-border pt-8">
				<h2 class="mb-3 font-serif text-base font-medium">Filed recently</h2>
				{#each filedGroups as g, gi (gi)}
					<h3 class="mt-5 mb-1 text-xs text-muted-foreground tabular-nums first:mt-0">{g.label}</h3>
					<ul class="text-sm lg:columns-2 lg:gap-x-10">
						{#each g.rows as f (f.symbol)}
							<li
								class="flex break-inside-avoid flex-wrap items-baseline gap-x-3 border-b border-border/50 py-2"
							>
								<a href="/company/{f.symbol}" class="font-serif hover:underline" title={f.name}
									>{f.symbol}</a
								>
								<span class="min-w-64 flex-1">{f.claim || f.name}</span>
							</li>
						{/each}
					</ul>
				{/each}
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
