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

	function fmt(value: number): string {
		return value.toLocaleString('en-NP', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
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
					value: `Rs ${formatLarge(shares * ltp)}`,
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
			empty: 'No gainers recorded for the latest session.',
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
			empty: 'No losers recorded for the latest session.',
			rows: (data.overview?.losers ?? []).map((m) => ({
				symbol: m.symbol,
				name: m.name,
				value: signed(m.changePercent),
				tone: tone(m.changePercent)
			}))
		},
		{
			title: 'Most traded',
			href: '/screener?sort=turnover',
			empty: 'No volume recorded for the latest session.',
			rows: mostTraded.map((r) => ({ ...r, tone: 'text-muted-foreground' }))
		},
		{
			title: 'Largest market cap',
			href: '/screener?sort=market_cap',
			empty: 'No listed-share counts on file yet.',
			rows: largestMarketCap.map((r) => ({ ...r, tone: 'text-foreground' }))
		}
	]);

	// Each card opens with the sentence its table supports (DESIGN.md: section anatomy).
	let cheapest = $derived(
		sectors.filter((s) => s.medianPe !== undefined).sort((a, b) => a.medianPe! - b.medianPe!)[0]
	);
	let mostProfitable = $derived(
		sectors
			.filter((s) => s.profitableShare !== undefined && s.quarterCount >= 5)
			.sort((a, b) => b.profitableShare! - a.profitableShare!)[0]
	);
	let filedProfits = $derived(filedRecently.filter((f) => f.claim.startsWith('Net profit')).length);
	let filedLosses = $derived(filedRecently.filter((f) => f.claim.startsWith('Net loss')).length);
	let topGainer = $derived(data.overview?.gainers?.[0]);
	let topLoser = $derived(data.overview?.losers?.[0]);

	// Session figures beside the index, in the order a NEPSE reader expects.
	let figures = $derived(
		marketDay
			? [
					{ label: 'Turnover', value: `Rs ${formatLarge(marketDay.turnover)}` },
					{ label: 'Shares traded', value: formatLarge(Number(marketDay.tradedShares)) },
					{ label: 'Transactions', value: Number(marketDay.transactions).toLocaleString('en-NP') },
					{ label: 'Scrips traded', value: marketDay.scripsTraded.toLocaleString('en-NP') },
					...(marketDay.marketCap > 0
						? [{ label: 'Market cap', value: `Rs ${formatLarge(marketDay.marketCap)}` }]
						: [])
				]
			: []
	);
</script>

<main class="mx-auto max-w-7xl px-4 py-6">
	<!-- The session: what the market did, how broadly, and the figures behind it -->
	<header class="card-report">
		<div class="grid gap-8 lg:grid-cols-[1fr_minmax(0,480px)]">
			<div class="min-w-0">
				<h1 class="font-serif text-3xl font-medium tracking-tight sm:text-4xl">NEPSE</h1>
				<p class="mt-1 text-base text-muted-foreground">
					Nepal Stock Exchange{#if marketDay}
						· <span class="tabular-nums">{marketDay.businessDateBs} · {marketDay.businessDate}</span
						>{/if}
				</p>
				<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">
					{#if marketDay}
						Closed at <span class="tabular-nums">{fmt(marketDay.indexClose)}</span>,
						<span class="tabular-nums {tone(marketDay.indexChange)}"
							>{word(marketDay.indexChange)}
							{Math.abs(marketDay.indexChangePercent).toFixed(2)}%</span
						>, on turnover of Rs {formatLarge(marketDay.turnover)}. {profitable} of {filed} companies
						with a filed quarter are profitable.
					{:else}
						{listed} listed companies in {sectors.length} sectors; {profitable} of {filed} with a filed
						quarter are profitable.
					{/if}
				</p>
				<p class="mt-6 text-xs text-muted-foreground">
					{#if marketDay}
						Session summary from NEPSE's close · Filings from company reports · Market cap is listed
						shares times the close
					{:else}
						No session summary recorded yet · Filings from company reports
					{/if}
				</p>
			</div>

			{#if marketDay}
				<div class="min-w-0 lg:pt-1">
					<p class="flex items-center gap-2 text-sm text-muted-foreground">
						<span class="size-2 rounded-full bg-primary" aria-hidden="true"></span> NEPSE index
					</p>
					<p class="mt-1 flex flex-wrap items-baseline gap-x-4">
						<span class="font-serif text-3xl font-medium tabular-nums"
							>{fmt(marketDay.indexClose)}</span
						>
						<span class="text-sm tabular-nums">
							<span class="text-muted-foreground">1D</span>
							<span class={tone(marketDay.indexChange)}>{signed(marketDay.indexChangePercent)}</span
							>
						</span>
					</p>

					<!-- Breadth: one bar, three segments, each named so colour is not the only carrier. -->
					{#if breadth > 0}
						<div
							class="mt-5 flex h-2 overflow-hidden rounded-full bg-muted"
							role="img"
							aria-label="{marketDay.advances} advanced, {marketDay.declines} declined, {marketDay.unchanged} unchanged"
						>
							<div class="bg-positive" style="width: {(marketDay.advances / breadth) * 100}%"></div>
							<div
								class="bg-negative {marketDay.advances > 0 ? 'border-l-2 border-card' : ''}"
								style="width: {(marketDay.declines / breadth) * 100}%"
							></div>
							<div
								class="bg-muted {marketDay.advances + marketDay.declines > 0
									? 'border-l-2 border-card'
									: ''}"
								style="width: {(marketDay.unchanged / breadth) * 100}%"
							></div>
						</div>
						<p class="mt-1.5 flex justify-between text-xs text-muted-foreground tabular-nums">
							<span><span class="text-positive">▲</span> {marketDay.advances} advanced</span>
							<span>{marketDay.unchanged} unchanged</span>
							<span>{marketDay.declines} declined <span class="text-negative">▼</span></span>
						</p>
					{/if}

					<dl class="mt-6 grid grid-cols-2 gap-x-6 gap-y-4 sm:grid-cols-4">
						{#each figures as f (f.label)}
							<div>
								<dt class="text-xs text-muted-foreground">{f.label}</dt>
								<dd class="mt-0.5 text-base font-medium tabular-nums">{f.value}</dd>
							</div>
						{/each}
					</dl>
				</div>
			{/if}
		</div>
	</header>

	<!-- The session's edges: who moved, who traded, who is largest -->
	<section id="movers" class="mt-6 card-report">
		<h2 class="font-serif text-2xl font-medium">Latest session</h2>
		<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">
			{#if topGainer && topLoser}
				{topGainer.symbol} led,
				<span class="text-positive tabular-nums">up {topGainer.changePercent.toFixed(2)}%</span>; {topLoser.symbol}
				fell furthest,
				<span class="text-negative tabular-nums"
					>down {Math.abs(topLoser.changePercent).toFixed(2)}%</span
				>.
			{:else}
				<span class="text-muted-foreground">No movers recorded for the latest session.</span>
			{/if}
		</p>
		<div class="mt-5 grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
			{#each movers as list (list.title)}
				<div>
					<div class="mb-2 flex items-baseline justify-between gap-2">
						<h3 class="font-sans text-sm font-medium">{list.title}</h3>
						{#if list.href}
							<a href={list.href} class="text-xs text-muted-foreground hover:underline">View all</a>
						{/if}
					</div>
					{#if list.rows.length === 0}
						<p class="text-xs text-muted-foreground">{list.empty}</p>
					{/if}
					<ul class="text-sm">
						{#each list.rows as r (r.symbol)}
							<li>
								<a
									href="/company/{r.symbol}"
									class="-mx-2 flex items-baseline justify-between gap-2 rounded-md px-2 py-1.5 hover:bg-accent/40"
									title={r.name}
								>
									<span class="font-medium">{r.symbol}</span>
									<span class="tabular-nums {r.tone}">{r.value}</span>
								</a>
							</li>
						{/each}
					</ul>
				</div>
			{/each}
		</div>
		{#if marketDay}
			<p class="mt-5 text-xs text-muted-foreground tabular-nums">
				Session of {marketDay.businessDateBs} · {marketDay.businessDate}. Market cap is listed
				shares times that close.
			</p>
		{/if}
	</section>

	<!-- Sectors: where value and profit sit -->
	{#if sectors.length > 0}
		<section id="sectors" class="mt-6 card-report">
			<h2 class="font-serif text-2xl font-medium">Sectors</h2>
			{#if cheapest || mostProfitable}
				<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">
					{#if cheapest && mostProfitable && cheapest.name === mostProfitable.name}
						{cheapest.name} is both the cheapest sector, at a median PE of
						<span class="tabular-nums">{ratio(cheapest.medianPe)}</span>, and the most profitable,
						with
						<span class="tabular-nums"
							>{Math.round((mostProfitable.profitableShare ?? 0) * 100)}%</span
						>
						of its {mostProfitable.quarterCount} filed companies in profit.
					{:else}
						{#if cheapest}
							{cheapest.name} is the cheapest sector at a median PE of
							<span class="tabular-nums">{ratio(cheapest.medianPe)}</span>{mostProfitable
								? ';'
								: '.'}
						{/if}
						{#if mostProfitable}
							{mostProfitable.name} is the most profitable, with
							<span class="tabular-nums"
								>{Math.round((mostProfitable.profitableShare ?? 0) * 100)}%</span
							>
							of its {mostProfitable.quarterCount} filed companies in profit.
						{/if}
					{/if}
				</p>
			{/if}
			<div class="table-scroll mt-5">
				<table class="w-full text-sm whitespace-nowrap tabular-nums">
					<thead class="text-left text-xs text-muted-foreground">
						<tr class="border-b border-border">
							<th class="sticky left-0 bg-card py-2 pr-4 font-normal">Sector</th>
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
							<tr class="border-b border-border/50 hover:bg-accent/40">
								<!-- Pinned so a phone can scroll the figures and keep the row's name. -->
								<td class="sticky left-0 bg-card py-2.5 pr-4">
									<a href="/screener?sector={s.sector}" class="hover:underline">{s.name}</a>
								</td>
								<td class="py-2.5 pr-4 text-right">{s.companyCount}</td>
								<td class="py-2.5 pr-4 text-right">
									{ratio(s.medianPe)}
									<span class="hidden text-xs text-muted-foreground sm:inline"
										>of {s.peCount} profitable</span
									>
								</td>
								<td class="py-2.5 pr-4 text-right">{ratio(s.medianPb)}</td>
								<td class="py-2.5 pr-4 text-right">
									<!-- The share as a bar beside the number, so the column reads without arithmetic. -->
									{#if s.profitableShare !== undefined}
										<span
											class="mr-2 inline-block h-1.5 w-16 overflow-hidden rounded-full bg-muted align-middle"
											aria-hidden="true"
										>
											<span
												class="block h-full bg-foreground/60"
												style="width: {s.profitableShare * 100}%"
											></span>
										</span>
									{/if}
									{s.profitableShare === undefined
										? '—'
										: `${Math.round(s.profitableShare * 100)}%`}
									<span class="text-xs text-muted-foreground">of {s.quarterCount}</span>
								</td>
								<td class="py-2.5 pr-4 text-right {tone(s.profitChangeYoy)}">
									{signed(s.profitChangeYoy, 1)}
								</td>
								<td class="py-2.5 text-right {tone(s.dayChange)}">{signed(s.dayChange)}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
			<p class="mt-3 text-xs text-muted-foreground">
				Medians over each company's latest quarter and latest close. Profit YoY sums year-to-date
				profit against the same quarter a year earlier. Today is the sector's median change.
			</p>
		</section>
	{/if}

	<!-- Filings: who reported, and the first thing their filing says -->
	{#if filedGroups.length > 0}
		<section id="filed" class="mt-6 card-report">
			<h2 class="font-serif text-2xl font-medium">Filed recently</h2>
			<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">
				{filedRecently.length} companies filed; {filedProfits} report a profit{filedLosses > 0
					? ` and ${filedLosses} a loss`
					: ''}.
			</p>
			{#each filedGroups as g, gi (gi)}
				<h3 class="mt-6 mb-1 font-sans text-xs font-normal text-muted-foreground tabular-nums">
					{g.label}
				</h3>
				<ul class="text-sm lg:columns-2 lg:gap-x-10">
					{#each g.rows as f (f.symbol)}
						<li
							class="flex break-inside-avoid flex-col gap-x-3 border-b border-border/50 py-2 sm:flex-row sm:items-baseline"
						>
							<a href="/company/{f.symbol}" class="font-medium hover:underline" title={f.name}
								>{f.symbol}</a
							>
							{#if f.source}
								<a
									href={f.source}
									target="_blank"
									rel="noreferrer"
									class="flex-1 underline decoration-dotted underline-offset-4 hover:text-primary"
									title="Open the filing">{f.claim || f.name}</a
								>
							{:else}
								<span class="flex-1 {f.claim ? '' : 'text-muted-foreground'}"
									>{f.claim || f.name}</span
								>
							{/if}
						</li>
					{/each}
				</ul>
			{/each}
		</section>
	{/if}
</main>
