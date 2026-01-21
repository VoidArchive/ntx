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
	import Search from '@lucide/svelte/icons/search';
	import Trash2 from '@lucide/svelte/icons/trash-2';
	import ArrowUp from '@lucide/svelte/icons/arrow-up';
	import ArrowDown from '@lucide/svelte/icons/arrow-down';
	import AlertTriangle from '@lucide/svelte/icons/alert-triangle';
	import CheckCircle from '@lucide/svelte/icons/check-circle';
	import Info from '@lucide/svelte/icons/info';
	import Banknote from '@lucide/svelte/icons/banknote';
	import { SectorChart } from '$lib/components/charts';
	import type { Portfolio, PortfolioSummary, Transaction, HealthTip } from '$lib/gen/ntx/v1/portfolio_pb';
	import type { Company } from '$lib/gen/ntx/v1/common_pb';

	
	let { data } = $props();
	const API_URL = data.apiUrl;
	
	// Companies from page load
	let companies = $derived<Company[]>(data.companies ?? []);

	let portfolios = $state<Portfolio[]>([]);
	let selectedPortfolio = $state<PortfolioSummary | null>(null);
	let isLoading = $state(true);
	let isLoadingSummary = $state(false);
	let showCreateModal = $state(false);
	let showAddTransactionModal = $state(false);
	let showTransactionsModal = $state(false);
	let transactionsForSymbol = $state('');
	let transactions = $state<Transaction[]>([]);
	let isLoadingTransactions = $state(false);
	let newPortfolioName = $state('');
	let error = $state<string | null>(null);

	// Analysis Data
	const SECTOR_COLORS = [
		'#f97316', '#3b82f6', '#10b981', '#ef4444', '#8b5cf6', '#eab308', '#ec4899', '#6366f1', '#14b8a6', '#f43f5e'
	];

	let sectorData = $derived(() => {
		if (!selectedPortfolio) return [];
		const sectorMap = new Map<string, number>();
		selectedPortfolio.holdings.forEach(h => {
			const sector = h.sector || 'Others';
			sectorMap.set(sector, (sectorMap.get(sector) || 0) + h.totalValue);
		});
		
		return Array.from(sectorMap.entries())
			.map(([label, value], i) => ({
				label,
				value,
				color: SECTOR_COLORS[i % SECTOR_COLORS.length]
			}))
			.sort((a, b) => b.value - a.value);
	});

	let dailyMovers = $derived(() => {
		if (!selectedPortfolio) return [];
		return [...selectedPortfolio.holdings]
			.filter(h => h.dayChangePercent !== 0)
			.sort((a, b) => Math.abs(b.dayChangePercent) - Math.abs(a.dayChangePercent))
			.slice(0, 4);
	});

	// Transaction form
	let txSymbol = $state('');
	let txSymbolSearch = $state('');
	let txType = $state<'BUY' | 'SELL'>('BUY');
	let txQuantity = $state(0);
	let txPrice = $state(0);
	let txDate = $state(new Date().toISOString().split('T')[0]);
	let showSymbolDropdown = $state(false);
	
	// Filtered companies for autocomplete
	let filteredCompanies = $derived(() => {
		if (!txSymbolSearch.trim()) return companies.slice(0, 10);
		const search = txSymbolSearch.toLowerCase();
		return companies
			.filter(c => 
				c.symbol.toLowerCase().includes(search) || 
				c.name.toLowerCase().includes(search)
			)
			.slice(0, 10);
	});
	
	function selectCompany(company: Company) {
		txSymbol = company.symbol;
		txSymbolSearch = company.symbol;
		showSymbolDropdown = false;
	}

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
			txSymbolSearch = '';
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
	
	async function showTransactionsFor(symbol: string) {
		if (!selectedPortfolio) return;
		transactionsForSymbol = symbol;
		showTransactionsModal = true;
		isLoadingTransactions = true;
		try {
			const response = await api.portfolio.listTransactions({
				portfolioId: selectedPortfolio.portfolioId,
				stockSymbol: symbol
			});
			transactions = response.transactions;
		} catch (err) {
			console.error('Failed to load transactions:', err);
			transactions = [];
		} finally {
			isLoadingTransactions = false;
		}
	}
	
	async function deleteTransaction(txId: bigint) {
		if (!confirm('Are you sure you want to delete this transaction?')) return;
		try {
			await api.portfolio.deleteTransaction({ transactionId: txId });
			// Reload transactions
			if (selectedPortfolio && transactionsForSymbol) {
				await showTransactionsFor(transactionsForSymbol);
			}
			// Reload summary to update holdings
			if (selectedPortfolio) {
				await loadPortfolioSummary(selectedPortfolio.portfolioId);
			}
		} catch (err) {
			console.error('Failed to delete transaction:', err);
		}
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
					<div class="rounded-xl border border-border bg-card/50 p-5 backdrop-blur-sm">
						<p class="text-xs text-muted-foreground">Est. Annual Yield</p>
						<div class="mt-1 flex items-center gap-2">
							<Banknote class="size-5 text-emerald-500" />
							<p class="text-2xl font-medium tabular-nums text-emerald-500">
								{formatCurrency(selectedPortfolio.projectedDividend ?? 0)}
							</p>
						</div>
					</div>
				</div>

				<!-- Analysis Grid -->
				<div class="mb-8 grid gap-8 lg:grid-cols-2">
					<!-- Daily Movers -->
					<div class="rounded-xl border border-border bg-card/50 p-5 backdrop-blur-sm">
						<h3 class="mb-4 font-serif text-lg font-medium">Daily Movers</h3>
						<div class="grid gap-3 sm:grid-cols-2">
							{#each dailyMovers() as mover (mover.stockSymbol)}
								<div class="flex items-center justify-between rounded-lg border border-border/50 bg-background/50 p-3">
									<div>
										<a href="/company/{mover.stockSymbol}" class="font-medium hover:text-primary hover:underline">
											{mover.stockSymbol}
										</a>
										<p class="text-xs text-muted-foreground">{formatCurrency(mover.currentPrice)}</p>
									</div>
									<div class="text-right">
										<p class="flex items-center justify-end gap-1 font-medium tabular-nums {mover.dayChangePercent >= 0 ? 'text-green-500' : 'text-red-500'}">
											{#if mover.dayChangePercent >= 0}
												<ArrowUp class="size-3" />
											{:else}
												<ArrowDown class="size-3" />
											{/if}
											{Math.abs(mover.dayChangePercent).toFixed(2)}%
										</p>
										<p class="text-xs text-muted-foreground tabular-nums">
											{mover.dayChangeValue >= 0 ? '+' : ''}{formatCurrency(mover.dayChangeValue)}
										</p>
									</div>
								</div>
							{/each}
							{#if dailyMovers().length === 0}
								<div class="col-span-2 py-8 text-center text-sm text-muted-foreground">
									No significant price movements today
								</div>
							{/if}
						</div>
					</div>

					<!-- Sector Allocation -->
					<div class="rounded-xl border border-border bg-card/50 p-5 backdrop-blur-sm">
						<h3 class="mb-4 font-serif text-lg font-medium">Sector Allocation</h3>
						{#if sectorData().length > 0}
							<SectorChart data={sectorData()} />
						{:else}
							<div class="py-12 text-center text-sm text-muted-foreground">
								No holdings to analyze
							</div>
						{/if}
					</div>
				</div>

				<!-- Recommendations / Health -->
				<div class="mb-8 rounded-xl border border-border bg-card/50 p-6 backdrop-blur-sm">
					<h3 class="mb-4 font-serif text-lg font-medium">Portfolio Health</h3>
					{#if !selectedPortfolio.healthTips || selectedPortfolio.healthTips.length === 0}
						<div class="flex items-center gap-3 rounded-lg border border-green-500/20 bg-green-500/10 p-4 text-green-600">
							<CheckCircle class="size-5 shrink-0" />
							<p class="text-sm font-medium">Your portfolio looks healthy! No immediate risks detected.</p>
						</div>
					{:else}
						<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
							{#each selectedPortfolio.healthTips as tip}
								<div class="flex items-start gap-3 rounded-lg border p-3 {tip.type === 'WARNING' ? 'border-amber-500/20 bg-amber-500/10 text-amber-700 dark:text-amber-400' : 'border-blue-500/20 bg-blue-500/10 text-blue-700 dark:text-blue-400'}">
									{#if tip.type === 'WARNING'}
										<AlertTriangle class="size-5 shrink-0 mt-0.5" />
									{:else}
										<Info class="size-5 shrink-0 mt-0.5" />
									{/if}
									<div>
										<a href="/company/{tip.symbol}" class="font-medium text-sm hover:underline">{tip.symbol}</a>
										<p class="text-xs opacity-90">{tip.message}</p>
									</div>
								</div>
							{/each}
						</div>
					{/if}
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
										<th class="px-4 py-3 text-right font-medium">Actions</th>
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
											<td class="px-4 py-3 text-right">
												<button
													onclick={() => showTransactionsFor(holding.stockSymbol)}
													class="rounded px-2 py-1 text-xs text-muted-foreground hover:bg-muted hover:text-foreground"
												>
													View
												</button>
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
				<div class="relative">
					<label class="mb-1 block text-sm font-medium">Symbol</label>
					<div class="relative">
						<Search class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
						<input
							type="text"
							bind:value={txSymbolSearch}
							onfocus={() => showSymbolDropdown = true}
							oninput={() => { showSymbolDropdown = true; txSymbol = ''; }}
							placeholder="Search companies..."
							class="w-full rounded-lg border border-border bg-background py-2.5 pl-10 pr-4 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
						/>
					</div>
					{#if txSymbol}
						<div class="mt-1 text-xs text-green-600">✓ Selected: {txSymbol}</div>
					{/if}
					{#if showSymbolDropdown && !txSymbol}
						<div class="absolute z-10 mt-1 max-h-48 w-full overflow-auto rounded-lg border border-border bg-card shadow-lg">
							{#each filteredCompanies() as company (company.symbol)}
								<button
									type="button"
									onclick={() => selectCompany(company)}
									class="flex w-full items-center gap-3 px-3 py-2 text-left text-sm hover:bg-muted"
								>
									<span class="font-medium">{company.symbol}</span>
									<span class="truncate text-xs text-muted-foreground">{company.name}</span>
								</button>
							{:else}
								<div class="px-3 py-2 text-sm text-muted-foreground">No companies found</div>
							{/each}
						</div>
					{/if}
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

<!-- Transaction History Modal -->
{#if showTransactionsModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
		<div class="w-full max-w-2xl rounded-2xl border border-border bg-card p-6 shadow-2xl">
			<div class="mb-4 flex items-center justify-between">
				<h2 class="font-serif text-xl font-medium">Transactions for {transactionsForSymbol}</h2>
				<button
					onclick={() => showTransactionsModal = false}
					class="text-muted-foreground hover:text-foreground"
				>✕</button>
			</div>
			
			{#if isLoadingTransactions}
				<div class="flex items-center justify-center py-8">
					<Loader2 class="size-6 animate-spin text-primary" />
				</div>
			{:else if transactions.length === 0}
				<div class="py-8 text-center text-sm text-muted-foreground">
					No transactions found
				</div>
			{:else}
				<div class="max-h-80 overflow-auto">
					<table class="w-full text-sm">
						<thead>
							<tr class="border-b border-border text-left text-xs text-muted-foreground">
								<th class="px-3 py-2 font-medium">Date</th>
								<th class="px-3 py-2 font-medium">Type</th>
								<th class="px-3 py-2 text-right font-medium">Qty</th>
								<th class="px-3 py-2 text-right font-medium">Price</th>
								<th class="px-3 py-2 text-right font-medium">Total</th>
								<th class="px-3 py-2 text-right font-medium"></th>
							</tr>
						</thead>
						<tbody>
							{#each transactions as tx (tx.id)}
								<tr class="border-b border-border/50 hover:bg-muted/50">
									<td class="px-3 py-2 tabular-nums">{tx.transactionDate}</td>
									<td class="px-3 py-2">
										<span class="rounded px-2 py-0.5 text-xs font-medium {tx.transactionType === 1 ? 'bg-green-500/10 text-green-600' : 'bg-red-500/10 text-red-600'}">
											{tx.transactionType === 1 ? 'BUY' : 'SELL'}
										</span>
									</td>
									<td class="px-3 py-2 text-right tabular-nums">{tx.quantity.toLocaleString()}</td>
									<td class="px-3 py-2 text-right tabular-nums">{tx.unitPrice.toFixed(2)}</td>
									<td class="px-3 py-2 text-right tabular-nums">{formatCurrency(Number(tx.quantity) * tx.unitPrice)}</td>
									<td class="px-3 py-2 text-right">
										<button
											onclick={() => deleteTransaction(tx.id)}
											class="rounded p-1 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
											title="Delete transaction"
										>
											<Trash2 class="size-4" />
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
			
			<div class="mt-4 flex justify-end">
				<button
					onclick={() => showTransactionsModal = false}
					class="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground"
				>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}

