<script lang="ts">
	import type { Company, Price } from '$lib/gen/ntx/v1/common_pb';
	import { Sector } from '$lib/gen/ntx/v1/common_pb';

	interface Props {
		company: Company;
		allCompanies: Company[];
		allPrices: Price[];
		class?: string;
	}

	let { company, allCompanies, allPrices, class: className = '' }: Props = $props();

	const sectorNames: Record<number, string> = {
		[Sector.COMMERCIAL_BANK]: 'Commercial Banking',
		[Sector.DEVELOPMENT_BANK]: 'Development Banking',
		[Sector.FINANCE]: 'Finance',
		[Sector.MICROFINANCE]: 'Microfinance',
		[Sector.LIFE_INSURANCE]: 'Life Insurance',
		[Sector.NON_LIFE_INSURANCE]: 'Non-Life Insurance',
		[Sector.HYDROPOWER]: 'Hydropower',
		[Sector.MANUFACTURING]: 'Manufacturing',
		[Sector.HOTEL]: 'Hotels & Tourism',
		[Sector.TRADING]: 'Trading',
		[Sector.INVESTMENT]: 'Investment',
		[Sector.MUTUAL_FUND]: 'Mutual Fund',
		[Sector.OTHERS]: 'Others'
	};

	function getPrice(companyId: bigint): Price | undefined {
		return allPrices?.find((p) => p.companyId === companyId);
	}

	function formatLarge(value: number): string {
		if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`;
		if (value >= 10_000_000) return `${(value / 10_000_000).toFixed(2)} Cr`;
		if (value >= 100_000) return `${(value / 100_000).toFixed(2)} L`;
		return value.toLocaleString();
	}

	function formatCurrency(value: number) {
		return new Intl.NumberFormat('en-NP', {
			style: 'currency',
			currency: 'NPR',
			maximumFractionDigits: 2
		}).format(value);
	}

	function formatPercent(value: number): string {
		const sign = value >= 0 ? '+' : '';
		return `${sign}${value.toFixed(2)}%`;
	}

	// Get ALL sector companies (including current company) with market cap
	let allSectorCompanies = $derived.by(() => {
		if (!allCompanies || !company.sector) return [];

		return allCompanies
			.filter((c) => c.sector === company.sector)
			.map((c) => {
				const p = getPrice(c.id);
				const shares = c.listedShares ? Number(c.listedShares) : 0;
				const ltp = p?.ltp ?? p?.close ?? 0;
				return {
					company: c,
					price: p,
					marketCap: shares * ltp,
					shares,
					ltp,
					isCurrent: c.id === company.id
				};
			})
			.filter((x) => x.marketCap > 0)
			.sort((a, b) => b.marketCap - a.marketCap);
	});

	// Peers are all sector companies except current
	let sectorPeers = $derived(allSectorCompanies.filter((x) => !x.isCurrent));

	// Top 10 peers by market cap
	let topPeers = $derived(sectorPeers.slice(0, 10));

	// Current company's rank (1-indexed position in the sorted list)
	let currentRank = $derived.by(() => {
		const idx = allSectorCompanies.findIndex((x) => x.isCurrent);
		return idx >= 0 ? idx + 1 : allSectorCompanies.length + 1;
	});
</script>

{#if sectorPeers.length >= 1}
	<div class={className}>
		<h3 class="mb-4 font-serif text-lg font-medium">Sector Peers</h3>
		<p class="mb-6 text-sm text-muted-foreground">
			{sectorNames[company.sector ?? Sector.OTHERS]} · {sectorPeers.length + 1} companies ·
			Rank #{currentRank} by market cap
		</p>

		<!-- Peer Table -->
		<div class="overflow-x-auto">
			<table class="w-full text-left text-sm">
				<thead class="bg-muted/50 text-xs uppercase text-muted-foreground">
					<tr>
						<th class="w-10 px-3 py-2.5 font-medium">#</th>
						<th class="px-3 py-2.5 font-medium">Symbol</th>
						<th class="px-3 py-2.5 text-right font-medium">LTP</th>
						<th class="px-3 py-2.5 text-right font-medium">Change</th>
						<th class="px-3 py-2.5 text-right font-medium">Market Cap</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-border">
					{#each topPeers as peer, i (peer.company.id)}
						{@const actualRank = i + 1 >= currentRank ? i + 2 : i + 1}
						<tr class="transition-colors hover:bg-muted/50">
							<td class="w-10 px-3 py-2.5 font-medium tabular-nums text-muted-foreground">
								{actualRank}
							</td>
							<td class="px-3 py-2.5">
								<a
									href="/company/{peer.company.symbol}"
									class="font-medium transition-colors hover:text-primary"
								>
									{peer.company.symbol}
								</a>
							</td>
							<td class="px-3 py-2.5 text-right tabular-nums">
								{formatCurrency(peer.ltp)}
							</td>
							<td
								class="px-3 py-2.5 text-right tabular-nums {(peer.price?.changePercent ?? 0) >= 0
									? 'text-positive'
									: 'text-negative'}"
							>
								{formatPercent(peer.price?.changePercent ?? 0)}
							</td>
							<td class="px-3 py-2.5 text-right font-medium tabular-nums">
								{formatLarge(peer.marketCap)}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		{#if sectorPeers.length > 10}
			<p class="mt-3 text-xs text-muted-foreground">
				Showing top 10 of {sectorPeers.length} peers by market cap
			</p>
		{/if}
	</div>
{/if}
