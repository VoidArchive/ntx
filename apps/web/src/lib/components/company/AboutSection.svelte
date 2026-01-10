<script lang="ts">
	import type { Company } from '$lib/gen/ntx/v1/common_pb';
	import { Sector, CompanyStatus } from '$lib/gen/ntx/v1/common_pb';
	import ExternalLink from '@lucide/svelte/icons/external-link';

	interface Props {
		company: Company;
	}

	let { company }: Props = $props();

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

	function getWebsiteUrl(website: string | undefined): string {
		if (!website) return '';
		return website.startsWith('http') ? website : `https://${website}`;
	}
</script>

<div class="text-sm">
	<h3 class="font-serif text-base font-medium">About {company.name}</h3>

	{#if company.website}
		<a
			href={getWebsiteUrl(company.website)}
			target="_blank"
			rel="noopener noreferrer"
			class="mt-2 inline-flex items-center gap-1 text-chart-1 hover:underline"
		>
			{company.website}
			<ExternalLink class="size-3" />
		</a>
	{/if}

	<p class="mt-3 leading-relaxed text-muted-foreground">
		{company.name} operates in the {sectorNames[company.sector ?? Sector.OTHERS]} sector on the Nepal
		Stock Exchange.
	</p>

	{#if company.status === CompanyStatus.ACTIVE}
		<p class="mt-3 text-xs text-positive">Active</p>
	{:else if company.status === CompanyStatus.SUSPENDED}
		<p class="mt-3 text-xs text-caution">Trading suspended</p>
	{:else if company.status === CompanyStatus.DELISTED}
		<p class="mt-3 text-xs text-negative">Delisted</p>
	{/if}
</div>
