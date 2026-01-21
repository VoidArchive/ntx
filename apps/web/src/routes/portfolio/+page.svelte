<script lang="ts">
	import { goto } from '$app/navigation';
	import { createApiClient } from '$lib/api/client';
	import { authStore } from '$lib/stores/auth.svelte';
	import Plus from '@lucide/svelte/icons/plus';
	import TrendingUp from '@lucide/svelte/icons/trending-up';
	import TrendingDown from '@lucide/svelte/icons/trending-down';
	import Briefcase from '@lucide/svelte/icons/briefcase';
	import Loader2 from '@lucide/svelte/icons/loader-2';
	import LogOut from '@lucide/svelte/icons/log-out';
	import type { Portfolio, PortfolioSummary } from '$lib/gen/ntx/v1/portfolio_pb';

	const API_URL = import.meta.env.DEV ? 'http://localhost:8080' : 'https://ntx-api.anishshrestha.com';

	let portfolios = $state<Portfolio[]>([]);
	let selectedPortfolio = $state<PortfolioSummary | null>(null);
	let isLoading = $state(true);
	let isLoadingSummary = $state(false);
	let showCreateModal = $state(false);
	let showAddTransactionModal = $state(false);
	let newPortfolioName = $state('');
	let error = $state<string | null>(null);

	// Transaction form
	let txSymbol = $state('');
	let txType = $state<'BUY' | 'SELL'>('BUY');
	let txQuantity = $state(0);
	let txPrice = $state(0);
	let txDate = $state(new Date().toISOString().split('T')[0]);

	// Redirect if not authenticated
	$effect(() => {
		if (!authStore.state.isAuthenticated) {
			goto('/login');
		}
	});

	$effect(() => {
		if (authStore.state.isAuthenticated) {
			loadPortfolios();
		}
	});

	const api = createApiClient(API_URL, () => authStore.getToken());

	async function loadPortfolios() {
		isLoading = true;
		error = null;
		try {
			const response = await api.portfolio.listPortfolios({});
			portfolios = response.portfolios;
			if (portfolios.length > 0 && !selectedPortfolio) {
				await loadPortfolioSummary(portfolios[0].id);
			}
		} catch (err) {
			error = 'Failed to load portfolios';
			console.error(err);
		} finally {
			isLoading = false;
		}
	}

	async function loadPortfolioSummary(portfolioId: bigint) {
		isLoadingSummary = true;
		try {
			const response = await api.portfolio.getPortfolioSummary({ portfolioId });
			selectedPortfolio = response.summary ?? null;
		} catch (err) {
			console.error('Failed to load portfolio summary:', err);
		} finally {
			isLoadingSummary = false;
		}
	}

	async function createPortfolio() {
		if (!newPortfolioName.trim()) return;
		try {
			await api.portfolio.createPortfolio({ name: newPortfolioName.trim() });
			newPortfolioName = '';
			showCreateModal = false;
			await loadPortfolios();
		} catch (err) {
			console.error('Failed to create portfolio:', err);
		}
	}

	async function addTransaction() {
		if (!selectedPortfolio || !txSymbol.trim() || txQuantity <= 0 || txPrice <= 0) return;
		try {
			await api.portfolio.addTransaction({
				portfolioId: selectedPortfolio.portfolioId,
				stockSymbol: txSymbol.toUpperCase().trim(),
				transactionType: txType === 'BUY' ? 1 : 2,
				quantity: BigInt(txQuantity),
				unitPrice: txPrice,
				transactionDate: txDate
			});
			// Reset form
			txSymbol = '';
			txQuantity = 0;
			txPrice = 0;
			showAddTransactionModal = false;
			// Reload summary
			await loadPortfolioSummary(selectedPortfolio.portfolioId);
		} catch (err) {
			console.error('Failed to add transaction:', err);
		}
	}

	function logout() {
		authStore.logout();
		goto('/login');
	}

	function formatCurrency(value: number): string {
		return `Rs. ${value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
	}

	function formatPercent(value: number): string {
		return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
	}
</script>

<svelte:head>
	<title>Portfolio - NTX</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<div class="min-h-screen bg-background">
	<div class="mx-auto max-w-7xl px-4 py-8">
		<!-- Header -->
		<div class="mb-8 flex items-center justify-between">
			<div>
				<h1 class="font-serif text-3xl font-medium">My Portfolio</h1>
				<p class="mt-1 text-muted-foreground">Track your investments and performance</p>
			</div>
			<button
				onclick={logout}
				class="flex items-center gap-2 rounded-lg border border-border px-4 py-2 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
			>
				<LogOut class="size-4" />
				Sign out
			</button>
		</div>

		{#if isLoading}
			<div class="flex items-center justify-center py-20">
				<Loader2 class="size-8 animate-spin text-primary" />
			</div>
		{:else if error}
			<div class="rounded-lg border border-destructive/50 bg-destructive/10 px-4 py-3 text-destructive">
				{error}
			</div>
		{:else if portfolios.length === 0}
			<!-- Empty state -->
			<div class="flex flex-col items-center justify-center rounded-2xl border border-dashed border-border bg-card/30 py-20">
				<Briefcase class="mb-4 size-12 text-muted-foreground/50" />
				<h2 class="text-lg font-medium">No portfolios yet</h2>
				<p class="mb-6 text-sm text-muted-foreground">Create your first portfolio to start tracking</p>
				<button
					onclick={() => (showCreateModal = true)}
					class="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"
				>
					<Plus class="size-4" />
					Create Portfolio
				</button>
			</div>
		{:else}
			<!-- Portfolio tabs -->
			<div class="mb-6 flex items-center gap-2 overflow-x-auto pb-2">
				{#each portfolios as portfolio (portfolio.id)}
					<button
						onclick={() => loadPortfolioSummary(portfolio.id)}
						class="shrink-0 rounded-lg px-4 py-2 text-sm font-medium transition-colors {selectedPortfolio?.portfolioId === portfolio.id
							? 'bg-primary text-primary-foreground'
							: 'bg-muted text-muted-foreground hover:bg-muted/80'}"
					>
						{portfolio.name}
					</button>
				{/each}
				<button
					onclick={() => (showCreateModal = true)}
					class="flex shrink-0 items-center gap-1 rounded-lg border border-dashed border-border px-4 py-2 text-sm text-muted-foreground transition-colors hover:border-foreground/50 hover:text-foreground"
				>
					<Plus class="size-4" />
					Add
				</button>
			</div>

			{#if isLoadingSummary}
				<div class="flex items-center justify-center py-20">
					<Loader2 class="size-6 animate-spin text-primary" />
				</div>
			{:else if selectedPortfolio}
				<!-- Summary cards -->
				<div class="mb-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
					<div class="rounded-xl border border-border bg-card/50 p-5 backdrop-blur-sm">
						<p class="text-xs text-muted-foreground">Total Invested</p>
						<p class="mt-1 text-2xl font-medium tabular-nums">{formatCurrency(selectedPortfolio.totalInvested)}</p>
					</div>
					<div class="rounded-xl border border-border bg-card/50 p-5 backdrop-blur-sm">
						<p class="text-xs text-muted-foreground">Current Value</p>
						<p class="mt-1 text-2xl font-medium tabular-nums">{formatCurrency(selectedPortfolio.totalCurrentValue)}</p>
					</div>
					<div class="rounded-xl border border-border bg-card/50 p-5 backdrop-blur-sm">
						<p class="text-xs text-muted-foreground">Total P/L</p>
						<p class="mt-1 flex items-center gap-2 text-2xl font-medium tabular-nums {selectedPortfolio.totalProfitLoss >= 0 ? 'text-green-500' : 'text-red-500'}">
							{#if selectedPortfolio.totalProfitLoss >= 0}
								<TrendingUp class="size-5" />
							{:else}
								<TrendingDown class="size-5" />
							{/if}
							{formatCurrency(Math.abs(selectedPortfolio.totalProfitLoss))}
						</p>
					</div>
					<div class="rounded-xl border border-border bg-card/50 p-5 backdrop-blur-sm">
						<p class="text-xs text-muted-foreground">P/L %</p>
						<p class="mt-1 text-2xl font-medium tabular-nums {selectedPortfolio.totalProfitLossPercent >= 0 ? 'text-green-500' : 'text-red-500'}">
							{formatPercent(selectedPortfolio.totalProfitLossPercent)}
						</p>
					</div>
				</div>

				<!-- Holdings table -->
				<div class="rounded-xl border border-border bg-card/50 backdrop-blur-sm">
					<div class="flex items-center justify-between border-b border-border p-4">
						<h2 class="font-medium">Holdings</h2>
						<button
							onclick={() => (showAddTransactionModal = true)}
							class="flex items-center gap-1 rounded-lg bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground"
						>
							<Plus class="size-3" />
							Add Transaction
						</button>
					</div>

					{#if selectedPortfolio.holdings.length === 0}
						<div class="py-12 text-center text-sm text-muted-foreground">
							No holdings yet. Add a transaction to get started.
						</div>
					{:else}
						<div class="overflow-x-auto">
							<table class="w-full text-sm">
								<thead>
									<tr class="border-b border-border text-left text-xs text-muted-foreground">
										<th class="px-4 py-3 font-medium">Symbol</th>
										<th class="px-4 py-3 text-right font-medium">Qty</th>
										<th class="px-4 py-3 text-right font-medium">Avg. Cost</th>
										<th class="px-4 py-3 text-right font-medium">LTP</th>
										<th class="px-4 py-3 text-right font-medium">Value</th>
										<th class="px-4 py-3 text-right font-medium">P/L</th>
										<th class="px-4 py-3 text-right font-medium">P/L %</th>
									</tr>
								</thead>
								<tbody>
									{#each selectedPortfolio.holdings as holding (holding.stockSymbol)}
										<tr class="border-b border-border/50 transition-colors hover:bg-muted/50">
											<td class="px-4 py-3">
												<a href="/company/{holding.stockSymbol}" class="font-medium hover:text-primary hover:underline">
													{holding.stockSymbol}
												</a>
											</td>
											<td class="px-4 py-3 text-right tabular-nums">{holding.quantity.toLocaleString()}</td>
											<td class="px-4 py-3 text-right tabular-nums">{holding.avgBuyPrice.toFixed(2)}</td>
											<td class="px-4 py-3 text-right tabular-nums">{holding.currentPrice.toFixed(2)}</td>
											<td class="px-4 py-3 text-right tabular-nums">{formatCurrency(holding.totalValue)}</td>
											<td class="px-4 py-3 text-right tabular-nums {holding.profitLoss >= 0 ? 'text-green-500' : 'text-red-500'}">
												{formatCurrency(holding.profitLoss)}
											</td>
											<td class="px-4 py-3 text-right tabular-nums {holding.profitLossPercent >= 0 ? 'text-green-500' : 'text-red-500'}">
												{formatPercent(holding.profitLossPercent)}
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				</div>
			{/if}
		{/if}
	</div>
</div>

<!-- Create Portfolio Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
		<div class="w-full max-w-md rounded-2xl border border-border bg-card p-6 shadow-2xl">
			<h2 class="mb-4 font-serif text-xl font-medium">Create Portfolio</h2>
			<input
				type="text"
				bind:value={newPortfolioName}
				placeholder="Portfolio name"
				class="mb-4 w-full rounded-lg border border-border bg-background px-4 py-2.5 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
			/>
			<div class="flex justify-end gap-2">
				<button
					onclick={() => (showCreateModal = false)}
					class="rounded-lg border border-border px-4 py-2 text-sm hover:bg-muted"
				>
					Cancel
				</button>
				<button
					onclick={createPortfolio}
					class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"
				>
					Create
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Add Transaction Modal -->
{#if showAddTransactionModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
		<div class="w-full max-w-md rounded-2xl border border-border bg-card p-6 shadow-2xl">
			<h2 class="mb-4 font-serif text-xl font-medium">Add Transaction</h2>
			<form onsubmit={(e) => { e.preventDefault(); addTransaction(); }} class="space-y-4">
				<div>
					<label class="mb-1 block text-sm font-medium">Symbol</label>
					<input
						type="text"
						bind:value={txSymbol}
						placeholder="e.g. NABIL"
						class="w-full rounded-lg border border-border bg-background px-4 py-2.5 text-sm uppercase focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
					/>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="mb-1 block text-sm font-medium">Type</label>
						<select
							bind:value={txType}
							class="w-full rounded-lg border border-border bg-background px-4 py-2.5 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
						>
							<option value="BUY">Buy</option>
							<option value="SELL">Sell</option>
						</select>
					</div>
					<div>
						<label class="mb-1 block text-sm font-medium">Date</label>
						<input
							type="date"
							bind:value={txDate}
							class="w-full rounded-lg border border-border bg-background px-4 py-2.5 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
						/>
					</div>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="mb-1 block text-sm font-medium">Quantity</label>
						<input
							type="number"
							bind:value={txQuantity}
							min="1"
							placeholder="100"
							class="w-full rounded-lg border border-border bg-background px-4 py-2.5 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
						/>
					</div>
					<div>
						<label class="mb-1 block text-sm font-medium">Price per unit</label>
						<input
							type="number"
							bind:value={txPrice}
							min="0.01"
							step="0.01"
							placeholder="1000.00"
							class="w-full rounded-lg border border-border bg-background px-4 py-2.5 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
						/>
					</div>
				</div>
				<div class="flex justify-end gap-2 pt-2">
					<button
						type="button"
						onclick={() => (showAddTransactionModal = false)}
						class="rounded-lg border border-border px-4 py-2 text-sm hover:bg-muted"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"
					>
						Add Transaction
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
