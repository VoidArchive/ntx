<script lang="ts">
	import { enhance } from '$app/forms';
	import type { Holding } from '#lib/gen/ntx/v1/portfolio_pb.js';
	import { HoldingStatus, TransactionKind } from '#lib/gen/ntx/v1/portfolio_pb.js';
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

	const today = new Date().toISOString().slice(0, 10);
	const rs = (v: number, digits = 2) =>
		v.toLocaleString('en-NP', { minimumFractionDigits: digits, maximumFractionDigits: digits });
	const n = (v: bigint | number) => Number(v).toLocaleString('en-NP');

	function href(h: Holding) {
		return h.isFund ? `/mutual-funds/${h.symbol}` : `/company/${h.symbol}`;
	}
	function asOf(h: Holding) {
		if (h.isFund) return data.navDates?.[h.symbol] ? `NAV, ${data.navDates[h.symbol]}` : '';
		return h.closeDay;
	}

	let editing = $state<string | undefined>();
	let draft = $derived(data.portfolio?.holdings.find((h) => h.symbol === editing));
</script>

<svelte:head>
	<title>Portfolio · NTX</title>
</svelte:head>

<div class="mx-auto min-h-screen max-w-7xl px-4 py-6">
	{#if data.needsToken}
		<header class="max-w-xl card-report">
			<h1 class="font-serif text-3xl font-medium tracking-tight">Portfolio</h1>
			<p class="mt-3 font-serif text-xl leading-snug">
				The owner's holdings. Enter the owner token once; it is kept in this browser only.
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

		<section class="mt-6 card-report" aria-label="Holdings">
			<h2 class="font-serif text-xl">Holdings</h2>
			<div class="-mx-6 mt-3 overflow-x-auto px-6 sm:mx-0 sm:px-0">
				<table class="w-full text-sm whitespace-nowrap tabular-nums [&_td+td]:pl-3 [&_th+th]:pl-3">
					<thead class="text-xs text-muted-foreground">
						<tr class="border-b border-border">
							<th class="py-2 text-left font-normal">Symbol</th>
							<th class="py-2 text-left font-normal">Sector</th>
							<th class="py-2 text-left font-normal">Status</th>
							<th class="py-2 text-right font-normal">Shares</th>
							<th class="py-2 text-right font-normal">WAC</th>
							<th class="py-2 text-right font-normal">Close</th>
							<th class="py-2 text-left font-normal">As of</th>
							<th class="py-2 text-right font-normal">Value</th>
							<th class="py-2 text-right font-normal">Weight</th>
							<th class="py-2 text-right font-normal">To build</th>
							<th class="py-2"></th>
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
									{statusLabel.get(h.status) ?? 'Hold'}
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
								><td colspan="11" class="py-12 text-center text-muted-foreground"
									>No holdings yet. Add one below.</td
								></tr
							>
						{/each}
					</tbody>
				</table>
			</div>
		</section>

		{#if draft}
			<section class="mt-6 card-report" aria-label="{draft.symbol} details">
				<h2 class="font-serif text-xl">{draft.symbol} · {draft.name}</h2>
				{#if draft.thesis}<p class="mt-2 max-w-3xl font-serif leading-snug">{draft.thesis}</p>{/if}

				<div class="mt-4 grid gap-6 lg:grid-cols-2">
					<div>
						<h3 class="text-xs text-muted-foreground">Transactions</h3>
						<table class="mt-1 w-full text-sm tabular-nums [&_td+td]:pl-3">
							<tbody>
								{#each draft.transactions as tx (tx.id)}
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
							<input type="hidden" name="symbol" value={draft.symbol} />
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
							{#each draft.notes as note (note.id)}
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
							<input type="hidden" name="symbol" value={draft.symbol} />
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

				<form method="POST" action="?/delete" use:enhance class="mt-6">
					<input type="hidden" name="symbol" value={draft.symbol} />
					<button type="submit" class="text-xs text-destructive underline underline-offset-2"
						>Remove {draft.symbol} and its rows</button
					>
				</form>
			</section>
		{/if}

		<section class="mt-6 card-report" aria-label="Add or edit a holding">
			<h2 class="font-serif text-xl">{draft ? `Edit ${draft.symbol}` : 'Add a holding'}</h2>
			<p class="mt-1 text-sm text-muted-foreground">
				A symbol with no shares is watched. A fund symbol is priced from the NAV snapshot.
			</p>
			{#key draft?.symbol}
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
							value={draft?.symbol ?? ''}
							readonly={!!draft}
							class="{field} uppercase"
						/>
					</label>
					<label class="flex flex-col gap-1">
						Status
						<select name="status" value={draft?.status ?? HoldingStatus.HOLD} class={field}>
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
							value={Number(draft?.targetShares ?? 0)}
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
							value={draft?.allocationPercent ?? 0}
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
							value={draft?.carryRs ?? 0}
							class={field}
						/>
					</label>
					<label class="flex flex-col gap-1">
						Review date
						<input type="date" name="review_date" value={draft?.reviewDate ?? ''} class={field} />
					</label>
					<label class="flex flex-col gap-1 sm:col-span-2">
						Frozen because
						<input name="frozen_reason" value={draft?.frozenReason ?? ''} class={field} />
					</label>
					<label class="flex flex-col gap-1 sm:col-span-2 lg:col-span-4">
						Thesis
						<textarea
							name="thesis"
							rows="2"
							class="rounded-md border border-input bg-background px-2 py-1 text-sm"
							>{draft?.thesis ?? ''}</textarea
						>
					</label>
					<label class="flex items-center gap-2">
						<input type="checkbox" name="drawdown_gate" checked={draft?.drawdownGate ?? false} />
						Drawdown gate applies
					</label>
					<div class="flex items-end">
						<button type="submit" class={button}>{draft ? 'Save' : 'Add'}</button>
					</div>
				</form>
			{/key}
		</section>
	{/if}
</div>
