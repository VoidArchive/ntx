<script lang="ts">
	import { goto } from '$app/navigation';
	import * as Command from '$lib/components/ui/command';
	import * as Dialog from '$lib/components/ui/dialog';
	import { company } from '$lib/api/client';
	import { getSectorName } from '$lib/utils/sector';
	import { debounce } from '$lib/utils/debounce';
	import { SearchIcon, Building2Icon, TrendingUpIcon, ClockIcon } from '@lucide/svelte';
	import type { Company } from '@ntx/api/ntx/v1/common_pb';

	let { open = $bindable(false) }: { open?: boolean } = $props();

	let query = $state('');
	let results = $state<Company[]>([]);
	let loading = $state(false);
	let recentSearches = $state<string[]>([]);

	// Load recent searches from localStorage
	$effect(() => {
		if (typeof window !== 'undefined') {
			const stored = localStorage.getItem('ntx-recent-searches');
			if (stored) {
				recentSearches = JSON.parse(stored);
			}
		}
	});

	// Keyboard shortcut handler
	$effect(() => {
		function handleKeydown(e: KeyboardEvent) {
			if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
				e.preventDefault();
				open = !open;
			}
		}

		if (typeof window !== 'undefined') {
			window.addEventListener('keydown', handleKeydown);
			return () => window.removeEventListener('keydown', handleKeydown);
		}
	});

	const searchCompanies = debounce(async (q: string) => {
		if (!q.trim()) {
			results = [];
			loading = false;
			return;
		}

		loading = true;
		try {
			const response = await company.listCompanies({ query: q });
			results = response.companies.slice(0, 10);
		} catch (err) {
			console.error('Search error:', err);
			results = [];
		} finally {
			loading = false;
		}
	}, 300);

	function onQueryChange(value: string) {
		query = value;
		if (value.trim()) {
			loading = true;
			searchCompanies(value);
		} else {
			results = [];
		}
	}

	function selectCompany(symbol: string) {
		// Add to recent searches
		const newRecent = [symbol, ...recentSearches.filter((s) => s !== symbol)].slice(0, 5);
		recentSearches = newRecent;
		if (typeof window !== 'undefined') {
			localStorage.setItem('ntx-recent-searches', JSON.stringify(newRecent));
		}

		open = false;
		query = '';
		results = [];
		goto(`/company/${symbol}`);
	}

	function clearRecentSearches() {
		recentSearches = [];
		if (typeof window !== 'undefined') {
			localStorage.removeItem('ntx-recent-searches');
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="overflow-hidden p-0 sm:max-w-lg">
		<Command.Root shouldFilter={false} class="[&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group-heading]]:text-muted-foreground [&_[cmdk-group]]:px-2 [&_[cmdk-input-wrapper]_svg]:h-5 [&_[cmdk-input-wrapper]_svg]:w-5 [&_[cmdk-input]]:h-12 [&_[cmdk-item]]:px-2 [&_[cmdk-item]]:py-3 [&_[cmdk-item]_svg]:h-5 [&_[cmdk-item]_svg]:w-5">
			<div class="flex items-center border-b px-3">
				<SearchIcon class="mr-2 h-4 w-4 shrink-0 opacity-50" />
				<Command.Input
					placeholder="Search companies by symbol or name..."
					class="flex h-11 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50"
					value={query}
					oninput={(e) => onQueryChange(e.currentTarget.value)}
				/>
			</div>
			<Command.List class="max-h-[400px] overflow-y-auto overflow-x-hidden">
				{#if loading}
					<Command.Loading class="py-6 text-center text-sm text-muted-foreground">
						Searching...
					</Command.Loading>
				{:else if query && results.length === 0}
					<Command.Empty class="py-6 text-center text-sm text-muted-foreground">
						No companies found.
					</Command.Empty>
				{:else if !query && recentSearches.length > 0}
					<Command.Group heading="Recent Searches">
						{#each recentSearches as symbol}
							<Command.Item
								value={symbol}
								onSelect={() => selectCompany(symbol)}
								class="flex cursor-pointer items-center gap-2"
							>
								<ClockIcon class="h-4 w-4 text-muted-foreground" />
								<span class="font-mono font-medium">{symbol}</span>
							</Command.Item>
						{/each}
						<button
							onclick={clearRecentSearches}
							class="w-full px-2 py-1.5 text-left text-xs text-muted-foreground hover:text-foreground"
						>
							Clear recent searches
						</button>
					</Command.Group>
				{:else if results.length > 0}
					<Command.Group heading="Companies">
						{#each results as result}
							<Command.Item
								value={result.symbol}
								onSelect={() => selectCompany(result.symbol)}
								class="flex cursor-pointer items-center justify-between gap-2"
							>
								<div class="flex items-center gap-3">
									<div
										class="flex h-8 w-8 items-center justify-center rounded-md bg-primary/10 text-xs font-bold text-primary"
									>
										{result.symbol.slice(0, 2)}
									</div>
									<div>
										<div class="font-mono font-medium">{result.symbol}</div>
										<div class="text-xs text-muted-foreground">{result.name}</div>
									</div>
								</div>
								<span
									class="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground"
								>
									{getSectorName(result.sector)}
								</span>
							</Command.Item>
						{/each}
					</Command.Group>
				{:else}
					<div class="py-6 text-center">
						<Building2Icon class="mx-auto h-8 w-8 text-muted-foreground/50" />
						<p class="mt-2 text-sm text-muted-foreground">
							Search for any NEPSE listed company
						</p>
						<p class="mt-1 text-xs text-muted-foreground/70">
							Try typing "NABIL" or "Nepal Bank"
						</p>
					</div>
				{/if}
			</Command.List>
		</Command.Root>
	</Dialog.Content>
</Dialog.Root>
