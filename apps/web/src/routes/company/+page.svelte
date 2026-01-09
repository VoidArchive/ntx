<script lang="ts">
	import CompanyCard from '$lib/components/CompanyCard.svelte';
	import { Sector } from '$lib/gen/ntx/v1/common_pb';

	let { data } = $props();

	// Filter state
	let selectedSector = $state<number | null>(null);
	let searchQuery = $state('');

	// Define sectors for filter pills
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

	// Filtered companies
	let filteredCompanies = $derived.by(() => {
		let companies = data.companies ?? [];

		// Filter by sector
		if (selectedSector !== null) {
			companies = companies.filter((c) => c.sector === selectedSector);
		}

		// Filter by search query
		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase();
			companies = companies.filter(
				(c) =>
					c.symbol.toLowerCase().includes(query) ||
					c.name.toLowerCase().includes(query)
			);
		}

		return companies;
	});

	function toggleSector(sector: number) {
		if (selectedSector === sector) {
			selectedSector = null;
		} else {
			selectedSector = sector;
		}
	}
</script>

<div class="screener">
	<header class="screener-header">
		<h1>NEPSE Companies</h1>
		<p class="subtitle">Find your next investment story</p>
	</header>

	<div class="filters">
		<input
			type="search"
			placeholder="Search by name or symbol..."
			bind:value={searchQuery}
			class="search-input"
		/>

		<div class="sector-pills">
			{#each sectors as sector}
				<button
					class="pill"
					class:active={selectedSector === sector.value}
					onclick={() => toggleSector(sector.value)}
				>
					{sector.label}
				</button>
			{/each}
		</div>
	</div>

	<p class="count">{filteredCompanies.length} companies</p>

	<div class="grid">
		{#each filteredCompanies as company (company.id)}
			<CompanyCard {company} />
		{/each}
	</div>

	{#if filteredCompanies.length === 0}
		<p class="no-results">No companies found matching your criteria.</p>
	{/if}
</div>

<style>
	.screener {
		max-width: 1200px;
		margin: 0 auto;
		padding: 2rem 1.5rem;
	}

	.screener-header {
		text-align: center;
		margin-bottom: 2rem;
	}

	.screener-header h1 {
		font-size: 2rem;
		margin: 0;
	}

	.subtitle {
		color: var(--muted-foreground);
		margin: 0.5rem 0 0;
	}

	.filters {
		margin-bottom: 1.5rem;
	}

	.search-input {
		width: 100%;
		padding: 0.75rem 1rem;
		font-size: 1rem;
		border: 1px solid var(--border);
		border-radius: var(--radius);
		background: var(--background);
		color: var(--foreground);
		margin-bottom: 1rem;
	}

	.search-input:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 2px var(--ring);
	}

	.sector-pills {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.pill {
		padding: 0.375rem 0.75rem;
		font-size: 0.875rem;
		border: 1px solid var(--border);
		border-radius: 999px;
		background: var(--background);
		color: var(--foreground);
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.pill:hover {
		border-color: var(--primary);
	}

	.pill.active {
		background: var(--primary);
		color: var(--primary-foreground);
		border-color: var(--primary);
	}

	.count {
		font-size: 0.875rem;
		color: var(--muted-foreground);
		margin-bottom: 1rem;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1rem;
	}

	.no-results {
		text-align: center;
		color: var(--muted-foreground);
		padding: 3rem;
	}
</style>
