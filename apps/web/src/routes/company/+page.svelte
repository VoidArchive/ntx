<script lang="ts">
	import type { PageData } from './$types';
	import type { ScreenerRow } from '#lib/gen/ntx/v1/company_pb.js';
	import { Sector } from '#lib/gen/ntx/v1/common_pb.js';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { SvelteSet, SvelteURLSearchParams } from 'svelte/reactivity';

	let { data }: { data: PageData } = $props();

	const sectors = [
		{ value: Sector.COMMERCIAL_BANK, label: 'Banks' },
		{ value: Sector.DEVELOPMENT_BANK, label: 'Dev Banks' },
		{ value: Sector.FINANCE, label: 'Finance' },
		{ value: Sector.MICROFINANCE, label: 'Microfinance' },
		{ value: Sector.LIFE_INSURANCE, label: 'Life Insurance' },
		{ value: Sector.NON_LIFE_INSURANCE, label: 'Non-Life Insurance' },
		{ value: Sector.HYDROPOWER, label: 'Hydropower' },
		{ value: Sector.MANUFACTURING, label: 'Manufacturing' },
		{ value: Sector.HOTEL, label: 'Hotels' },
		{ value: Sector.TRADING, label: 'Trading' },
		{ value: Sector.INVESTMENT, label: 'Investment' },
		{ value: Sector.MUTUAL_FUND, label: 'Mutual Fund' },
		{ value: Sector.OTHERS, label: 'Others' }
	];
	const sectorLabel = new Map(sectors.map((s) => [s.value, s.label]));

	type NumericKey = 'close' | 'pe' | 'pb' | 'eps' | 'yield' | 'marketCap' | 'turnover';
	type SortKey = 'symbol' | 'rating' | NumericKey;

	// Column order; `param` is the URL spelling, `desc` the direction a first click sorts by.
	const columns: { key: SortKey; param: string; label: string; desc: boolean }[] = [
		{ key: 'symbol', param: 'symbol', label: 'Symbol', desc: false },
		{ key: 'close', param: 'close', label: 'Close', desc: true },
		{ key: 'pe', param: 'pe', label: 'PE', desc: false },
		{ key: 'pb', param: 'pb', label: 'PB', desc: false },
		{ key: 'eps', param: 'eps', label: 'EPS', desc: true },
		{ key: 'yield', param: 'yield', label: 'Yield', desc: true },
		{ key: 'marketCap', param: 'market_cap', label: 'Mkt cap, Rs cr', desc: true },
		{ key: 'turnover', param: 'turnover', label: 'Turnover, Rs cr', desc: true },
		{ key: 'rating', param: 'rating', label: 'Rating', desc: false }
	];

	// The URL is the only filter state, so a filter set is a link and back works.
	let params = $derived(page.url.searchParams);
	let q = $derived(params.get('q') ?? '');
	let sectorFilter = $derived(
		new SvelteSet((params.get('sector') ?? '').split(',').filter(Boolean).map(Number))
	);
	let peMin = $derived(params.get('pe_min') ?? '');
	let peMax = $derived(params.get('pe_max') ?? '');
	let pbMin = $derived(params.get('pb_min') ?? '');
	let pbMax = $derived(params.get('pb_max') ?? '');
	let minRating = $derived(params.get('rating') ?? '');
	let sortCol = $derived(
		columns.find((c) => c.param === params.get('sort')) ??
			columns.find((c) => c.key === 'marketCap')!
	);
	let sortDesc = $derived(params.has('dir') ? params.get('dir') === 'desc' : sortCol.desc);
	let hasFilters = $derived(
		!!(q || sectorFilter.size || peMin || peMax || pbMin || pbMax || minRating)
	);

	function setParams(changes: Record<string, string>) {
		const next = new SvelteURLSearchParams(page.url.searchParams.toString());
		for (const [k, v] of Object.entries(changes)) {
			if (v) next.set(k, v);
			if (!v) next.delete(k);
		}
		goto(`?${next}`, { replace: true, shallow: true });
	}

	function toggleSector(value: number) {
		const next = new SvelteSet(sectorFilter);
		if (next.has(value)) next.delete(value);
		if (!next.has(value) && !sectorFilter.has(value)) next.add(value);
		setParams({ sector: [...next].join(',') });
	}

	function sortBy(col: (typeof columns)[number]) {
		const desc = col.key === sortCol.key ? !sortDesc : col.desc;
		setParams({ sort: col.param, dir: desc ? 'desc' : 'asc' });
	}

	function clear() {
		goto('?', { replace: true, shallow: true });
	}

	function inRange(value: number | undefined, min: string, max: string): boolean {
		if (!min && !max) return true;
		if (value === undefined) return false;
		if (min && value < Number(min)) return false;
		if (max && value > Number(max)) return false;
		return true;
	}

	let rows = $derived.by(() => {
		const needle = q.trim().toLowerCase();
		const out = data.rows.filter((r) => {
			if (
				needle &&
				!r.symbol.toLowerCase().includes(needle) &&
				!r.name.toLowerCase().includes(needle)
			)
				return false;
			if (sectorFilter.size && !sectorFilter.has(r.sector)) return false;
			if (!inRange(r.pe, peMin, peMax)) return false;
			if (!inRange(r.pb, pbMin, pbMax)) return false;
			// Letters order A < B < C < D; unrated never meets a minimum.
			if (minRating && (!r.rating || r.rating > minRating)) return false;
			return true;
		});
		const key = sortCol.key;
		const dir = sortDesc ? -1 : 1;
		return out.sort((a, b) => {
			if (key === 'symbol') return dir * a.symbol.localeCompare(b.symbol);
			if (key === 'rating') return dir * rank(a.rating).localeCompare(rank(b.rating));
			// Missing values sort last whichever way the column points.
			const av = a[key];
			const bv = b[key];
			if (av === undefined || bv === undefined)
				return Number(av === undefined) - Number(bv === undefined);
			return dir * (av - bv);
		});
	});

	function rank(rating: string): string {
		return rating || (sortDesc ? '' : 'Z');
	}

	function num(value: number | undefined, digits = 2): string {
		if (value === undefined) return '—';
		return value.toLocaleString('en-NP', {
			minimumFractionDigits: digits,
			maximumFractionDigits: digits
		});
	}

	function crore(value: number | undefined, digits = 0): string {
		return value === undefined ? '—' : num(value / 1e7, digits);
	}

	function cell(r: ScreenerRow, key: SortKey): string {
		if (key === 'symbol') return r.symbol;
		if (key === 'rating') return r.rating || '—';
		if (key === 'close') return num(r.close);
		if (key === 'pe' || key === 'pb') return r[key] === undefined ? '—' : `${num(r[key], 1)}x`;
		if (key === 'eps') return num(r.eps);
		if (key === 'yield') return r.yield === undefined ? '—' : `${num(r.yield, 1)}%`;
		if (key === 'marketCap') return crore(r.marketCap);
		return crore(r.turnover, 2);
	}

	const inputClass =
		'h-8 w-20 rounded-md border border-input bg-background px-2 text-sm tabular-nums placeholder:text-muted-foreground focus:border-foreground focus:outline-none';
</script>

<svelte:head>
	<title>Companies · NTX</title>
</svelte:head>

<div class="mx-auto min-h-screen max-w-7xl px-4 py-8">
	<header>
		<h1 class="font-serif text-3xl font-medium tracking-tight">Companies</h1>
		<p class="mt-1 text-sm text-muted-foreground">
			{data.rows.length} listed on NEPSE. Ratios from each company's latest quarterly filing at its latest
			close; the rating is one letter for its eight checks.
		</p>
	</header>

	<!-- Filters: every control writes the URL -->
	<div class="mt-6 flex flex-col gap-3 text-sm">
		<div class="flex flex-wrap items-center gap-x-4 gap-y-2">
			<input
				type="search"
				placeholder="Symbol or name"
				value={q}
				oninput={(e) => setParams({ q: e.currentTarget.value })}
				class="{inputClass} w-48"
			/>
			<label class="flex items-center gap-1">
				PE
				<input
					type="number"
					min="0"
					placeholder="min"
					value={peMin}
					oninput={(e) => setParams({ pe_min: e.currentTarget.value })}
					class={inputClass}
				/>
				<input
					type="number"
					min="0"
					placeholder="max"
					value={peMax}
					oninput={(e) => setParams({ pe_max: e.currentTarget.value })}
					class={inputClass}
				/>
			</label>
			<label class="flex items-center gap-1">
				PB
				<input
					type="number"
					min="0"
					step="0.1"
					placeholder="min"
					value={pbMin}
					oninput={(e) => setParams({ pb_min: e.currentTarget.value })}
					class={inputClass}
				/>
				<input
					type="number"
					min="0"
					step="0.1"
					placeholder="max"
					value={pbMax}
					oninput={(e) => setParams({ pb_max: e.currentTarget.value })}
					class={inputClass}
				/>
			</label>
			<label class="flex items-center gap-1">
				Rating at least
				<select
					value={minRating}
					onchange={(e) => setParams({ rating: e.currentTarget.value })}
					class="{inputClass} w-16"
				>
					<option value="">any</option>
					{#each ['A', 'B', 'C', 'D'] as letter (letter)}
						<option value={letter}>{letter}</option>
					{/each}
				</select>
			</label>
			{#if hasFilters}
				<button
					onclick={clear}
					class="text-muted-foreground underline underline-offset-2 hover:text-foreground"
				>
					Clear
				</button>
			{/if}
		</div>
		<div class="flex flex-wrap gap-1.5">
			{#each sectors as sector (sector.value)}
				<button
					onclick={() => toggleSector(sector.value)}
					aria-pressed={sectorFilter.has(sector.value)}
					class="rounded-full px-2.5 py-1 text-xs transition-colors {sectorFilter.has(sector.value)
						? 'bg-foreground text-background'
						: 'bg-muted text-muted-foreground hover:bg-accent hover:text-foreground'}"
				>
					{sector.label}
				</button>
			{/each}
		</div>
	</div>

	<p class="mt-4 text-xs text-muted-foreground tabular-nums">
		{rows.length} of {data.rows.length} companies
	</p>

	<div class="mt-2 overflow-x-auto">
		<table class="w-full text-sm whitespace-nowrap tabular-nums [&_td+td]:pl-3 [&_th+th]:pl-3">
			<thead class="text-xs text-muted-foreground">
				<tr class="border-b border-border">
					{#each columns as col (col.key)}
						<th
							class="py-2 font-normal {col.key === 'symbol' ? 'text-left' : 'text-right'}"
							aria-sort={col.key === sortCol.key
								? sortDesc
									? 'descending'
									: 'ascending'
								: undefined}
						>
							<button
								onclick={() => sortBy(col)}
								class="hover:text-foreground {col.key === sortCol.key ? 'text-foreground' : ''}"
							>
								{col.label}{#if col.key === sortCol.key}
									<span aria-hidden="true">{sortDesc ? ' ▼' : ' ▲'}</span>{/if}
							</button>
						</th>
						{#if col.key === 'symbol'}
							<th class="py-2 text-left font-normal">Name</th>
							<th class="py-2 text-left font-normal">Sector</th>
							<th class="py-2 text-left font-normal">As of</th>
						{/if}
					{/each}
				</tr>
			</thead>
			<tbody>
				{#each rows as r (r.symbol)}
					<tr class="border-b border-border/50 hover:bg-muted/50">
						{#each columns as col (col.key)}
							<td class="py-1.5 {col.key === 'symbol' ? 'text-left' : 'text-right'}">
								{#if col.key === 'symbol'}
									<a href="/company/{r.symbol}" class="font-serif font-medium hover:underline"
										>{r.symbol}</a
									>
								{:else}
									{cell(r, col.key)}
								{/if}
							</td>
							{#if col.key === 'symbol'}
								<td class="max-w-44 truncate py-1.5 text-muted-foreground" title={r.name}
									>{r.name}</td
								>
								<td class="py-1.5 text-muted-foreground">{sectorLabel.get(r.sector) ?? 'Others'}</td
								>
								<!-- Each row's own filing and close: rows differ, so one footer line would not do. -->
								<td class="py-1.5 text-xs text-muted-foreground"
									>{r.period || 'not recorded'} · {r.closeDate || 'not recorded'}</td
								>
							{/if}
						{/each}
					</tr>
				{:else}
					<tr
						><td colspan={columns.length + 3} class="py-12 text-center text-muted-foreground"
							>No company matches. <button onclick={clear} class="underline">Clear filters</button
							></td
						></tr
					>
				{/each}
			</tbody>
		</table>
	</div>
</div>
