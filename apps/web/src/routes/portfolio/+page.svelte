<script lang="ts">
	import { enhance } from '$app/forms';
	import type { Candidate, Gate, Holding } from '#lib/gen/ntx/v1/portfolio_pb.js';
	import { HoldingStatus, TransactionKind } from '#lib/gen/ntx/v1/portfolio_pb.js';
	import GoalChart from '#lib/components/charts/GoalChart.svelte';
	import BankMetricForm from '#lib/components/portfolio/BankMetricForm.svelte';
	import CapBars from '#lib/components/portfolio/CapBars.svelte';
	import Gates from '#lib/components/portfolio/Gates.svelte';
	import {
		VERDICT,
		byVerdict,
		figure,
		isEvent,
		verdictOf
	} from '#lib/components/portfolio/gates.js';
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	const statuses = [
		{ value: HoldingStatus.BUILD, label: 'Build' },
		{ value: HoldingStatus.HOLD, label: 'Hold' },
		{ value: HoldingStatus.FROZEN, label: 'Frozen' }
	];
	const statusLabel = new Map(statuses.map((s) => [s.value, s.label]));

	// Past-tense labels: the words that recommend have no place on the page.
	const kinds = [
		{ value: TransactionKind.BUY, label: 'Bought' },
		{ value: TransactionKind.SELL, label: 'Sold' },
		{ value: TransactionKind.BONUS, label: 'Bonus' },
		{ value: TransactionKind.RIGHTS, label: 'Rights' },
		{ value: TransactionKind.IPO, label: 'IPO' }
	];
	const kindLabel = new Map(kinds.map((k) => [k.value, k.label]));

	const field = 'h-8 rounded-md border border-input bg-background px-2 text-sm';
	const button = 'h-8 rounded-md border border-input bg-background px-3 text-sm hover:bg-accent';
	const table = 'w-full text-sm whitespace-nowrap tabular-nums [&_td+td]:pl-3 [&_th+th]:pl-3';
	const th = 'py-2 font-normal';
	const summary =
		'cursor-pointer font-serif text-xl marker:text-muted-foreground [&::-webkit-details-marker]:hidden';

	const today = new Date().toISOString().slice(0, 10);
	const rs = (v: number, digits = 2) =>
		v.toLocaleString('en-NP', { minimumFractionDigits: digits, maximumFractionDigits: digits });
	const n = (v: bigint | number) => Number(v).toLocaleString('en-NP');
	const lakh = (v: number) => (v >= 1e5 ? `${(v / 1e5).toFixed(2)} lakh` : rs(v, 0));
	const crore = (v: number | undefined) =>
		v === undefined ? 'not recorded' : `Rs ${(v / 1e7).toFixed(2)} crore`;

	function href(h: Holding) {
		return h.isFund ? `/mutual-funds/${h.symbol}` : `/company/${h.symbol}`;
	}
	function asOf(h: Holding) {
		if (h.isFund) return data.navDates?.[h.symbol] ? `NAV, ${data.navDates[h.symbol]}` : '';
		return h.closeDay;
	}
	// The quarter a needs-data sentence names, to prefill the form: "needs Q1 2083/84 NPL …".
	function wanted(c: Candidate) {
		const m = /needs (Q[1-4]) (\d{4}\/\d{2})/.exec(c.reason);
		return m ? { quarter: m[1], fiscalYear: m[2] } : {};
	}

	let editing = $state<string | undefined>();
	let entering = $state<string | undefined>();
	let lookup = $state('');
	let selected = $derived(data.portfolio?.holdings.find((h) => h.symbol === editing));
	// Every gate across holdings, Triggered and Needs data first; event gates last with their tick date.
	let gates = $derived(
		(data.portfolio?.holdings ?? []).flatMap((h) => h.gates).sort(byVerdict) as Gate[]
	);
	let triggered = $derived(gates.filter((g) => g.verdict === 'triggered'));
	let gatesNeeding = $derived(gates.filter((g) => g.verdict === 'needs data'));
	let caps = $derived([...(data.portfolio?.caps ?? [])].sort(byVerdict));
	let rules = $derived(data.portfolio?.rules);
	let sheet = $derived(data.portfolio?.orderSheet);
	let goal = $derived(data.portfolio?.goal);
	let calendar = $derived(data.portfolio?.calendar ?? []);
	let nextDate = $derived(calendar.find((e) => !e.overdue));
	let candidates = $derived(data.portfolio?.candidates ?? []);
	let passes = $derived(candidates.filter((c) => c.verdict === 'passes'));
	let needing = $derived(candidates.filter((c) => c.verdict === 'needs data'));
	let away = $derived(
		candidates.filter(
			(c) => c.verdict === 'turned away' && c.symbol.includes(lookup.trim().toUpperCase())
		)
	);
	let numberRules = $derived<[string, string, number][][]>(
		rules
			? [
					[
						['soft_cap_percent', 'Soft cap %', rules.softCapPercent],
						['hard_cap_percent', 'Hard cap %', rules.hardCapPercent],
						['sector_cap_percent', 'Sector cap %', rules.sectorCapPercent],
						['max_positions', 'Max positions', rules.maxPositions]
					],
					[
						['min_entry_rs', 'Minimum entry (Rs)', rules.minEntryRs],
						[
							'max_pe_to_sector_median',
							'PE at most, times the sector median',
							rules.maxPeToSectorMedian
						]
					],
					[
						['npl_ceiling_percent', 'NPL ceiling %', rules.nplCeilingPercent],
						['car_floor_percent', 'CAR floor %', rules.carFloorPercent]
					],
					[
						['goal_rs', 'Goal (Rs)', rules.goalRs],
						['monthly_budget_rs', 'Monthly budget (Rs)', rules.monthlyBudgetRs],
						['fund_sip_rs', 'Fund SIP (Rs)', rules.fundSipRs],
						['multiplier_below_median', 'Multiplier below the median', rules.multiplierBelowMedian],
						[
							'multiplier_bottom_quarter',
							'Multiplier in the bottom quarter',
							rules.multiplierBottomQuarter
						]
					]
				]
			: []
	);
	const ruleGroups = ['Caps', 'Candidates', 'Lenders', 'Goal and budget'];
</script>

<svelte:head>
	<title>Portfolio · NTX</title>
</svelte:head>

{#snippet candidateRows(list: Candidate[])}
	{#each list as c (c.symbol)}
		<tr class="border-b border-border/50 hover:bg-accent/40">
			<td class="py-1.5">
				<a href="/company/{c.symbol}" class="font-medium hover:underline">{c.symbol}</a>
				<span class="ml-2 text-muted-foreground">{c.name}</span>
				{#if c.behind}<span class="block text-xs text-muted-foreground">{c.behind}</span>{/if}
			</td>
			<td class="py-1.5 text-muted-foreground">{c.sector}</td>
			<td class="py-1.5">{c.rating}</td>
			<td class="py-1.5 text-right">
				{rs(c.close)}
				<span class="block text-xs text-muted-foreground">{c.closeDay || 'not recorded'}</span>
			</td>
			<td class="py-1.5 text-right">
				{#if c.pe === undefined}not recorded{:else}{rs(c.pe, 1)}x
					<span class="block text-xs text-muted-foreground">{c.pePeriod}</span>{/if}
				{#if c.annualizedPe !== undefined}
					<span class="block">{rs(c.annualizedPe, 1)}x</span>
					<span class="block text-xs text-muted-foreground">on {c.annualizedPeriod}</span>
				{/if}
			</td>
			<td class="py-1.5 text-right text-muted-foreground"
				>{c.sectorPeMedian === undefined ? 'not recorded' : `${rs(c.sectorPeMedian, 1)}x`}</td
			>
			<td class="max-w-xs py-1.5 text-xs whitespace-normal text-muted-foreground">
				{#if c.metricsPeriod}
					<span class="text-sm text-foreground"
						>NPL {figure(c.nplPercent, '%')} · CAR {figure(c.carPercent, '%')} · {crore(
							c.distributableProfit
						)}</span
					>
					<a href={c.metricsSource} rel="noreferrer" class="block underline">{c.metricsPeriod}</a>
					{#if c.reason}<span class="block">{c.reason}</span>{/if}
				{:else}
					{c.reason || c.basin || c.note}
				{/if}
			</td>
			<td class="py-1.5 text-right">{n(c.shares)}</td>
			<td class="py-1.5 text-right">{rs(c.costRs, 0)}</td>
			<td class="py-1.5 text-right">{rs(c.months, 1)}</td>
		</tr>
	{/each}
{/snippet}

<div class="mx-auto min-h-screen max-w-7xl px-4 py-6">
	{#if data.needsToken}
		<header class="max-w-xl card-report">
			<h1 class="font-serif text-3xl font-medium tracking-tight">Portfolio</h1>
			<p class="mt-3 font-serif text-xl leading-snug">
				The owner's holdings. Enter the owner token once; this browser keeps it as a cookie and
				sends it only to NTX.
			</p>
			<form method="POST" action="?/token" use:enhance class="mt-6 flex gap-2">
				<input
					type="password"
					name="token"
					required
					autocomplete="off"
					aria-label="Owner token"
					class="{field} flex-1"
				/>
				<button type="submit" class={button}>Open</button>
			</form>
			{#if form?.message}<p class="mt-2 text-sm text-destructive">{form.message}</p>{/if}
		</header>
	{:else if data.portfolio}
		{@const p = data.portfolio}
		<header class="card-report">
			<div class="flex items-baseline justify-between gap-4">
				<h1 class="font-serif text-3xl font-medium tracking-tight">Portfolio</h1>
				<form method="POST" action="?/forget" use:enhance>
					<button type="submit" class="text-xs text-muted-foreground underline underline-offset-2"
						>Forget token</button
					>
				</form>
			</div>
			<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">{p.opening}</p>
			{#if form?.message}<p class="mt-2 text-sm text-destructive">{form.message}</p>{/if}
		</header>

		<!-- The monthly read: what to buy, what waits on the owner, the next date, the goal. -->
		<section class="mt-6 card-report" aria-label="This month">
			<div class="grid gap-8 lg:grid-cols-[1fr_20rem]">
				<div>
					<h2 class="font-serif text-xl">This month</h2>
					{#if sheet}
						<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">{sheet.opening}</p>
						<form method="POST" action="?/executed" use:enhance class="table-scroll mt-3">
							<table class={table}>
								<thead class="text-xs text-muted-foreground">
									<tr class="border-b border-border">
										<th class="{th} text-left">Symbol</th>
										<th class="{th} text-right">Available</th>
										<th class="{th} text-right">Close</th>
										<th class="{th} text-right">Shares</th>
										<th class="{th} text-right">Cost</th>
										<th class="{th} text-right">Carry out</th>
										<th class="{th} text-left">Note</th>
									</tr>
								</thead>
								<tbody>
									{#each sheet.lines as l (l.symbol)}
										<tr class="border-b border-border/50">
											<td class="py-1.5 font-medium">{l.symbol}</td>
											<td class="py-1.5 text-right">{l.availableRs ? rs(l.availableRs, 0) : '—'}</td
											>
											<td class="py-1.5 text-right">
												{l.close === undefined ? 'not recorded' : rs(l.close)}
												<span class="block text-xs text-muted-foreground">{l.closeDay}</span>
											</td>
											<td class="py-1.5 text-right">{l.shares ? n(l.shares) : '—'}</td>
											<td class="py-1.5 text-right">{l.costRs ? rs(l.costRs, 0) : '—'}</td>
											<td class="py-1.5 text-right">
												<input
													type="number"
													name="carry:{l.symbol}"
													min="0"
													step="1"
													value={Math.round(l.newCarryRs)}
													aria-label="{l.symbol} carry out"
													class="{field} w-24 text-right"
												/>
											</td>
											<td class="py-1.5 text-xs text-muted-foreground">{l.note}</td>
										</tr>
									{:else}
										<tr
											><td colspan="7" class="py-3 text-muted-foreground"
												>No Build or Frozen holding: nothing to propose this month.</td
											></tr
										>
									{/each}
								</tbody>
							</table>
							<p class="mt-2 text-xs text-muted-foreground">
								A proposal, never an order placed: whole lots of ten at the latest close, the
								remainder carried. Edit a carry before marking the month executed.
							</p>
							{#if sheet.lines.length > 0}
								<button type="submit" class="{button} mt-3">Mark {sheet.month} executed</button>
							{/if}
						</form>
					{/if}
				</div>
				<dl class="grid gap-5 text-sm sm:grid-cols-3 lg:grid-cols-1">
					<div>
						<dt class="text-xs text-muted-foreground">Triggered or waiting on you</dt>
						<dd class="mt-1 font-serif text-3xl font-medium tabular-nums">
							{triggered.length + gatesNeeding.length + needing.length}
						</dd>
						<dd class="mt-1 text-muted-foreground">
							{#each triggered as g (g.id)}
								<a href="#gates" class="block text-foreground">{g.symbol} {g.label}</a>
							{/each}
							{#if gatesNeeding.length}
								<a href="#gates" class="block"
									>{gatesNeeding.length === 1
										? '1 gate needs'
										: `${gatesNeeding.length} gates need`} data</a
								>
							{/if}
							{#if needing.length}
								<a href="#candidates" class="block"
									>{needing.length === 1
										? '1 candidate needs'
										: `${needing.length} candidates need`} data</a
								>
							{/if}
						</dd>
					</div>
					<div>
						<dt class="text-xs text-muted-foreground">Next date</dt>
						<dd class="mt-1 font-serif text-3xl font-medium tabular-nums">
							{nextDate?.day ?? '—'}
						</dd>
						<dd class="mt-1 text-muted-foreground">
							{#if nextDate}<a href="#calendar">{nextDate.symbol} {nextDate.what}</a
								>{:else}{p.calendarOpening}{/if}
						</dd>
					</div>
					{#if goal}
						<div>
							<dt class="text-xs text-muted-foreground">Goal</dt>
							<dd class="mt-1 font-serif text-3xl font-medium tabular-nums">
								Rs {lakh(goal.valueRs)}
							</dd>
							<dd class="mt-1 text-muted-foreground">
								<a href="#goal" class="block">of Rs {lakh(goal.goalRs)}</a>
								{#if goal.flatMonth}<span class="block">{goal.flatMonth} at flat prices</span>{/if}
								{#if goal.returnMonth}<span class="block"
										>{goal.returnMonth} at the realised return</span
									>{/if}
							</dd>
						</div>
					{/if}
				</dl>
			</div>
		</section>

		{#if goal}
			<section id="goal" class="mt-6 card-report" aria-label="Goal">
				<h2 class="font-serif text-xl">Goal</h2>
				<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">{goal.opening}</p>
				{#if goal.series.length > 1 || goal.band.length > 1}
					<GoalChart {goal} class="mt-4 h-56" />
					<p class="mt-2 text-xs text-muted-foreground">
						Value at each month's last close, the thinner line what was paid in; fund units at what
						they cost until the latest month, since NAV history is not stored. The band runs from
						flat prices to the realised return.
					</p>
				{/if}
			</section>
		{/if}

		<section id="holdings" class="mt-6 card-report" aria-label="Holdings">
			<h2 class="font-serif text-xl">Holdings</h2>
			{#if rules}
				<div class="mt-4"><CapBars holdings={p.holdings} {rules} /></div>
			{/if}
			<div class="table-scroll mt-6">
				<table class={table}>
					<thead class="text-xs text-muted-foreground">
						<tr class="border-b border-border">
							<th class="{th} text-left">Symbol</th>
							<th class="{th} text-left">Sector</th>
							<th class="{th} text-left">Status</th>
							<th class="{th} text-right">Shares</th>
							<th class="{th} text-right">WAC</th>
							<th class="{th} text-right">Close</th>
							<th class="{th} text-left">As of</th>
							<th class="{th} text-right">Value</th>
							<th class="{th} text-right">Weight</th>
							<th class="{th} text-left">Cap</th>
							<th class="{th} text-right">To build</th>
							<th class={th}></th>
						</tr>
					</thead>
					<tbody>
						{#each p.holdings as h (h.symbol)}
							<tr class="border-b border-border/50 hover:bg-accent/40">
								<td class="py-1.5">
									<a href={href(h)} class="font-medium hover:underline">{h.symbol}</a>
									<span class="ml-2 text-muted-foreground">{h.name}</span>
								</td>
								<td class="py-1.5 text-muted-foreground">{h.sector}</td>
								<td class="py-1.5">
									{statusLabel.get(h.status)}
									{#if h.status === HoldingStatus.FROZEN && h.frozenReason}
										<span class="text-xs text-muted-foreground"
											>· {h.frozenReason}{h.reviewDate ? `, review ${h.reviewDate}` : ''}</span
										>
									{/if}
								</td>
								<td class="py-1.5 text-right">{h.quantity ? n(h.quantity) : 'watched'}</td>
								<td class="py-1.5 text-right">{h.quantity ? rs(h.wac) : '—'}</td>
								<td class="py-1.5 text-right"
									>{h.close === undefined ? 'not recorded' : rs(h.close)}</td
								>
								<td class="py-1.5 text-xs text-muted-foreground">{asOf(h) || 'not recorded'}</td>
								<td class="py-1.5 text-right">{h.close === undefined ? '—' : rs(h.value, 0)}</td>
								<td class="py-1.5 text-right"
									>{h.close === undefined ? '—' : `${rs(h.weightPercent, 1)}%`}</td
								>
								<td class="py-1.5 text-xs">{VERDICT[h.capVerdict] ?? '—'}</td>
								<td class="py-1.5 text-right text-muted-foreground">
									{h.targetShares > 0n ? n(h.targetShares) : '—'}
								</td>
								<td class="py-1.5 text-right">
									<button
										type="button"
										class="text-xs text-muted-foreground underline underline-offset-2"
										onclick={() => (editing = editing === h.symbol ? undefined : h.symbol)}
										>{editing === h.symbol ? 'Close' : 'Open'}</button
									>
								</td>
							</tr>
						{:else}
							<tr
								><td colspan="12" class="py-12 text-center text-muted-foreground"
									>No holdings yet. Add one below.</td
								></tr
							>
						{/each}
					</tbody>
				</table>
			</div>
			<ul class="mt-3 space-y-1 text-sm">
				{#each caps as c (c.rule + c.detail)}
					<li>
						<span class="font-medium">{VERDICT[c.verdict] ?? c.verdict}</span>
						<span class="text-muted-foreground"> · {c.rule}: {c.detail}</span>
					</li>
				{/each}
			</ul>
		</section>

		<section id="gates" class="mt-6 card-report" aria-label="Gates">
			<h2 class="font-serif text-xl">Gates</h2>
			<p class="mt-1 text-sm text-muted-foreground">
				Your written conditions against the latest filings. Each verdict names the period it read.
			</p>
			<div class="table-scroll mt-3">
				<table class={table}>
					<thead class="text-xs text-muted-foreground">
						<tr class="border-b border-border">
							<th class="{th} text-left">Holding</th>
							<th class="{th} text-left">Gate</th>
							<th class="{th} text-right">Value</th>
							<th class="{th} text-left">Period</th>
							<th class="{th} text-left">Verdict</th>
						</tr>
					</thead>
					<tbody>
						{#each gates as g (g.id)}
							<tr class="border-b border-border/50">
								<td class="py-1.5 font-medium">{g.symbol}</td>
								<td class="py-1.5">{g.label}</td>
								<td class="py-1.5 text-right">
									{#if !isEvent(g)}
										{figure(g.value, g.unit)}{#if g.prior !== undefined}
											<span class="text-xs text-muted-foreground">
												vs {figure(g.prior, g.unit)}</span
											>{/if}
									{/if}
								</td>
								<td class="py-1.5 text-xs text-muted-foreground">
									{#if isEvent(g)}
										checked {g.checkedOn || 'never'}
									{:else if g.source}<a href={g.source} class="underline" rel="noreferrer"
											>{g.period}</a
										>{:else}{g.period || 'no period'}{/if}
								</td>
								<td class="py-1.5 font-medium">{verdictOf(g)}</td>
							</tr>
						{:else}
							<tr
								><td colspan="5" class="py-3 text-muted-foreground"
									>No gates written. Open a holding above to write one.</td
								></tr
							>
						{/each}
					</tbody>
				</table>
			</div>
		</section>

		<section id="candidates" class="mt-6 card-report" aria-label="Candidates">
			<h2 class="font-serif text-xl">Candidates</h2>
			<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">{p.candidatesOpening}</p>
			<p class="mt-1 text-sm text-muted-foreground">
				What your rules admit from the screener, not a ranking: not held, rated at least the
				minimum, the dividend record holding, sector headroom after a minimum entry, PE within the
				sector rule, no second hydro in a basin you hold, and for a lender your NPL, CAR and
				distributable-profit rules read from the figures you typed. Entry is the smallest whole lot
				that reaches the minimum.
			</p>
			<div class="table-scroll mt-3">
				<table class={table}>
					<thead class="text-xs text-muted-foreground">
						<tr class="border-b border-border">
							<th class="{th} text-left">Symbol</th>
							<th class="{th} text-left">Sector</th>
							<th class="{th} text-left">Rating</th>
							<th class="{th} text-right">Close</th>
							<th class="{th} text-right">PE, trailing and latest</th>
							<th class="{th} text-right">Sector median</th>
							<th class="{th} text-left">Basin · lender figures</th>
							<th class="{th} text-right">Entry</th>
							<th class="{th} text-right">Cost</th>
							<th class="{th} text-right">Months</th>
						</tr>
					</thead>
					<tbody>
						{@render candidateRows(passes)}
						{#if passes.length === 0}
							<tr
								><td colspan="10" class="py-3 text-muted-foreground"
									>No company passes the rules at the latest close.</td
								></tr
							>
						{/if}
					</tbody>
				</table>
			</div>

			{#if needing.length}
				<h3 class="mt-6 font-serif text-lg">Needs data</h3>
				<p class="mt-1 text-sm text-muted-foreground">
					Lenders every other rule admits, waiting on the figures from their latest filing. None
					passes until its quarter is typed with the filing's date and URL.
				</p>
				<div class="table-scroll mt-3">
					<table class={table}>
						<thead class="text-xs text-muted-foreground">
							<tr class="border-b border-border">
								<th class="{th} text-left">Symbol</th>
								<th class="{th} text-left">Sector</th>
								<th class="{th} text-left">Rating</th>
								<th class="{th} text-right">Close</th>
								<th class="{th} text-right">PE, trailing and latest</th>
								<th class="{th} text-right">Sector median</th>
								<th class="{th} text-left">Missing</th>
								<th class="{th} text-right">Entry</th>
								<th class="{th} text-right">Cost</th>
								<th class="{th} text-right">Months</th>
							</tr>
						</thead>
						<tbody>
							{#each needing as c (c.symbol)}
								{@render candidateRows([c])}
								<tr class="border-b border-border/50">
									<td colspan="10" class="py-1">
										<button
											type="button"
											class="text-xs text-muted-foreground underline underline-offset-2"
											onclick={() => (entering = entering === c.symbol ? undefined : c.symbol)}
											>{entering === c.symbol
												? 'Close'
												: `Enter ${wanted(c).quarter ?? ''} ${wanted(c).fiscalYear ?? ''} for ${c.symbol}`}</button
										>
										{#if entering === c.symbol}
											<BankMetricForm symbol={c.symbol} {...wanted(c)} />
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}

			<details class="mt-6">
				<summary class="cursor-pointer text-sm text-muted-foreground"
					>Turned away: {candidates.filter((c) => c.verdict === 'turned away').length} companies, each
					with the rule that turned it away</summary
				>
				<input
					type="search"
					bind:value={lookup}
					placeholder="Symbol"
					aria-label="Find a symbol"
					class="{field} mt-3 w-40 uppercase"
				/>
				<ul class="mt-3 space-y-1 text-sm">
					{#each away as c (c.symbol)}
						<li class="grid grid-cols-[4.5rem_1fr] gap-2 sm:grid-cols-[4.5rem_10rem_2rem_1fr]">
							<a href="/company/{c.symbol}" class="font-medium hover:underline">{c.symbol}</a>
							<span class="hidden text-muted-foreground sm:block">{c.sector}</span>
							<span class="hidden sm:block">{c.rating}</span>
							<span class="text-muted-foreground">{c.reason}</span>
						</li>
					{:else}
						<li class="text-muted-foreground">No such symbol among the turned away.</li>
					{/each}
				</ul>
			</details>
		</section>

		<section id="calendar" class="mt-6 card-report" aria-label="Calendar">
			<h2 class="font-serif text-xl">Calendar</h2>
			<p class="mt-3 max-w-3xl font-serif text-xl leading-snug">{p.calendarOpening}</p>
			<p class="mt-1 text-sm text-muted-foreground">
				Filing due dates (quarter end plus thirty days), review dates and event gates due a
				re-check. No notices.
			</p>
			<ul class="mt-3 space-y-1 text-sm tabular-nums">
				{#each calendar as e (e.day + e.symbol + e.what)}
					<li class="flex gap-3 {e.overdue ? 'text-destructive' : ''}">
						<span class="w-24 shrink-0">{e.day}</span>
						<span class="w-16 shrink-0 font-medium">{e.symbol}</span>
						<span class="text-muted-foreground">{e.what}</span>
					</li>
				{:else}
					<li class="text-muted-foreground">Nothing dated.</li>
				{/each}
			</ul>
		</section>

		{#if selected}
			<section class="mt-6 card-report" aria-label="{selected.symbol} details">
				<h2 class="font-serif text-xl">{selected.symbol} · {selected.name}</h2>
				{#if selected.thesis}<p class="mt-2 max-w-3xl font-serif leading-snug">
						{selected.thesis}
					</p>{/if}

				<div class="mt-4 grid gap-6 lg:grid-cols-2">
					<div>
						<h3 class="text-xs text-muted-foreground">Transactions</h3>
						<table class="mt-1 w-full text-sm tabular-nums [&_td+td]:pl-3">
							<tbody>
								{#each selected.transactions as tx (tx.id)}
									<tr class="border-b border-border/50">
										<td class="py-1">{tx.day}</td>
										<td class="py-1">{kindLabel.get(tx.kind)}</td>
										<td class="py-1 text-right">{n(tx.quantity)}</td>
										<td class="py-1 text-right">{rs(tx.price)}</td>
									</tr>
								{:else}
									<tr
										><td class="py-1 text-muted-foreground"
											>None recorded; this is a watched company.</td
										></tr
									>
								{/each}
							</tbody>
						</table>
						<form
							method="POST"
							action="?/transaction"
							use:enhance
							class="mt-3 flex flex-wrap items-center gap-2 text-sm"
						>
							<input type="hidden" name="symbol" value={selected.symbol} />
							<input type="date" name="day" value={today} required class={field} />
							<select name="kind" class={field}>
								{#each kinds as k (k.value)}<option value={k.value}>{k.label}</option>{/each}
							</select>
							<input
								type="number"
								name="quantity"
								min="1"
								step="1"
								required
								aria-label="Quantity"
								class="{field} w-24"
							/>
							<input
								type="number"
								name="price"
								min="0"
								step="0.01"
								required
								aria-label="Price per share"
								class="{field} w-28"
							/>
							<button type="submit" class={button}>Record</button>
						</form>
					</div>

					<div>
						<h3 class="text-xs text-muted-foreground">Notes</h3>
						<ul class="mt-1 space-y-1 text-sm">
							{#each selected.notes as note (note.id)}
								<li>
									<span class="text-muted-foreground tabular-nums">{note.day}</span>
									{note.text}
								</li>
							{:else}
								<li class="text-muted-foreground">No notes yet.</li>
							{/each}
						</ul>
						<form
							method="POST"
							action="?/note"
							use:enhance
							class="mt-3 flex flex-wrap gap-2 text-sm"
						>
							<input type="hidden" name="symbol" value={selected.symbol} />
							<input type="date" name="day" value={today} required class={field} />
							<input
								type="text"
								name="text"
								required
								aria-label="Note"
								class="{field} min-w-64 flex-1"
							/>
							<button type="submit" class={button}>Add</button>
						</form>
					</div>
				</div>

				{#if !selected.isFund}
					<div class="mt-6">
						<Gates symbol={selected.symbol} gates={selected.gates} />
					</div>
				{/if}

				<form method="POST" action="?/delete" use:enhance class="mt-6">
					<input type="hidden" name="symbol" value={selected.symbol} />
					<button type="submit" class="text-xs text-destructive underline underline-offset-2"
						>Remove {selected.symbol} and its rows</button
					>
				</form>
			</section>
		{/if}

		<!-- The forms sit below the monthly read, closed until needed. -->
		<details class="mt-6 card-report" open={!!selected}>
			<summary class={summary}>{selected ? `Edit ${selected.symbol}` : 'Add a holding'}</summary>
			<p class="mt-1 text-sm text-muted-foreground">
				A symbol with no shares is watched. A fund symbol is priced from the NAV snapshot.
			</p>
			{#key selected?.symbol}
				<form
					method="POST"
					action="?/holding"
					use:enhance
					class="mt-3 grid gap-3 text-sm sm:grid-cols-2 lg:grid-cols-4"
				>
					<label class="flex flex-col gap-1">
						Symbol
						<input
							name="symbol"
							required
							value={selected?.symbol ?? ''}
							readonly={!!selected}
							class="{field} uppercase"
						/>
					</label>
					<label class="flex flex-col gap-1">
						Status
						<select name="status" value={selected?.status ?? HoldingStatus.HOLD} class={field}>
							{#each statuses as s (s.value)}<option value={s.value}>{s.label}</option>{/each}
						</select>
					</label>
					<label class="flex flex-col gap-1">
						Shares to build to
						<input
							type="number"
							name="target_shares"
							min="0"
							step="1"
							value={Number(selected?.targetShares ?? 0)}
							class={field}
						/>
					</label>
					<label class="flex flex-col gap-1">
						Monthly allocation %
						<input
							type="number"
							name="allocation_percent"
							min="0"
							max="100"
							step="1"
							value={selected?.allocationPercent ?? 0}
							class={field}
						/>
					</label>
					<label class="flex flex-col gap-1">
						Carry (Rs)
						<input
							type="number"
							name="carry_rs"
							min="0"
							step="1"
							value={selected?.carryRs ?? 0}
							class={field}
						/>
					</label>
					<label class="flex flex-col gap-1">
						Review date
						<input
							type="date"
							name="review_date"
							value={selected?.reviewDate ?? ''}
							class={field}
						/>
					</label>
					<label class="flex flex-col gap-1 sm:col-span-2">
						Frozen because
						<input name="frozen_reason" value={selected?.frozenReason ?? ''} class={field} />
					</label>
					<label class="flex flex-col gap-1 sm:col-span-2 lg:col-span-4">
						Thesis
						<textarea
							name="thesis"
							rows="2"
							class="rounded-md border border-input bg-background px-2 py-1 text-sm"
							>{selected?.thesis ?? ''}</textarea
						>
					</label>
					{#if selected}
						<label class="flex items-center gap-2">
							<input type="hidden" name="drawdown_gate_known" value="1" />
							<input type="checkbox" name="drawdown_gate" checked={selected.drawdownGate} />
							Drawdown gate applies
						</label>
					{:else}
						<p class="self-center text-xs text-muted-foreground">
							The drawdown gate starts on for hydropower and off elsewhere; edit after adding.
						</p>
					{/if}
					<div class="flex items-end">
						<button type="submit" class={button}>{selected ? 'Save' : 'Add'}</button>
					</div>
				</form>
			{/key}
		</details>

		{#if rules}
			<details class="mt-6 card-report">
				<summary class={summary}>Rules</summary>
				<p class="mt-1 text-sm text-muted-foreground">
					The caps, the candidate and lender rules, the goal and the monthly budget the sections
					above read. One row, yours to edit. A ceiling, floor or multiple of zero switches that
					rule off.
				</p>
				<form
					method="POST"
					action="?/rules"
					use:enhance
					class="mt-3 grid gap-6 text-sm sm:grid-cols-2 lg:grid-cols-4"
				>
					{#each numberRules as group, i (ruleGroups[i])}
						<div class="flex flex-col gap-3">
							<h3 class="text-xs font-semibold tracking-wide text-muted-foreground uppercase">
								{ruleGroups[i]}
							</h3>
							{#each group as [name, text, value] (name)}
								<label class="flex flex-col gap-1">
									{text}
									<input type="number" {name} step="any" min="0" {value} class={field} />
								</label>
							{/each}
							{#if ruleGroups[i] === 'Candidates'}
								<label class="flex flex-col gap-1">
									Rating at least
									<select
										name="candidate_min_rating"
										value={rules.candidateMinRating}
										class={field}
									>
										{#each ['A', 'B', 'C', 'D'] as letter (letter)}<option value={letter}
												>{letter}</option
											>{/each}
									</select>
								</label>
								<label class="flex items-center gap-2">
									<input
										type="checkbox"
										name="one_hydro_per_basin"
										checked={rules.oneHydroPerBasin}
									/>
									One hydro per basin
								</label>
								<label class="flex items-center gap-2">
									<input
										type="checkbox"
										name="candidate_dividend_record"
										checked={rules.candidateDividendRecord}
									/>
									The dividend record must hold
								</label>
							{:else if ruleGroups[i] === 'Lenders'}
								<label class="flex items-center gap-2">
									<input
										type="checkbox"
										name="distributable_above_zero"
										checked={rules.distributableAboveZero}
									/>
									Distributable profit above zero
								</label>
							{/if}
						</div>
					{/each}
					<div class="flex items-end">
						<button type="submit" class={button}>Save rules</button>
					</div>
				</form>
			</details>
		{/if}

		<details class="mt-6 card-report">
			<summary class={summary}>Export and import</summary>
			<p class="mt-1 text-sm text-muted-foreground">
				One JSON file of every holding, transaction, note, gate, rule and bank figure. Importing a
				file replaces them all; nothing is kept back. No scheduled backup: download one when the
				book changes.
			</p>
			<div class="mt-3 flex flex-wrap items-center gap-3 text-sm">
				<a href="/portfolio/export" download class="{button} inline-flex items-center"
					>Download the file</a
				>
				<form
					method="POST"
					action="?/import"
					enctype="multipart/form-data"
					use:enhance
					class="flex flex-wrap items-center gap-2"
				>
					<input type="file" name="file" accept="application/json,.json" required class={field} />
					<button type="submit" class={button}>Replace with this file</button>
				</form>
			</div>
			{#if form?.message}<p class="mt-2 text-sm text-destructive">{form.message}</p>{/if}
		</details>
	{/if}
</div>
