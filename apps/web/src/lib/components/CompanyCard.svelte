<script lang="ts">
	import type { Company, Price } from '$lib/gen/ntx/v1/common_pb';
	import { Sector } from '$lib/gen/ntx/v1/common_pb';

	interface Props {
		company: Company;
		price?: Price;
		miniStory?: string;
	}

	let { company, price, miniStory }: Props = $props();

	function formatPrice(value: number | undefined): string {
		if (value === undefined) return '—';
		return value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	function formatChange(value: number | undefined): string {
		if (value === undefined) return '';
		const prefix = value > 0 ? '+' : '';
		return `${prefix}${value.toFixed(2)}%`;
	}

	const sectorLabels: Record<number, string> = {
		[Sector.COMMERCIAL_BANK]: 'Banking',
		[Sector.DEVELOPMENT_BANK]: 'Dev Bank',
		[Sector.FINANCE]: 'Finance',
		[Sector.MICROFINANCE]: 'Microfinance',
		[Sector.LIFE_INSURANCE]: 'Life Insurance',
		[Sector.NON_LIFE_INSURANCE]: 'Non-Life Insurance',
		[Sector.HYDROPOWER]: 'Hydropower',
		[Sector.MANUFACTURING]: 'Manufacturing',
		[Sector.HOTEL]: 'Hotels',
		[Sector.TRADING]: 'Trading',
		[Sector.INVESTMENT]: 'Investment',
		[Sector.MUTUAL_FUND]: 'Mutual Fund',
		[Sector.OTHERS]: 'Others'
	};
</script>

<a
	href="/company/{company.symbol}"
	class="group block border-b border-border py-4 transition-colors hover:bg-muted/30"
>
	<div class="flex items-start justify-between gap-4">
		<!-- Left: Symbol & Name -->
		<div class="min-w-0 flex-1">
			<div class="flex items-baseline gap-2">
				<span class="font-serif text-lg tracking-tight group-hover:underline">
					{company.symbol}
				</span>
				<span class="text-xs text-muted-foreground">
					{sectorLabels[company.sector ?? Sector.OTHERS] ?? 'Others'}
				</span>
			</div>
			<p class="mt-0.5 truncate text-sm text-muted-foreground">
				{company.name}
			</p>
		</div>

		<!-- Right: Price -->
		{#if price?.ltp}
			<div class="text-right">
				<div class="flex items-baseline gap-2">
					<span class="text-lg font-medium tabular-nums">
						{formatPrice(price.ltp)}
					</span>
					{#if price.changePercent !== undefined}
						<span
							class="text-sm tabular-nums {price.changePercent > 0
								? 'text-positive'
								: price.changePercent < 0
									? 'text-negative'
									: 'text-muted-foreground'}"
						>
							{formatChange(price.changePercent)}
						</span>
					{/if}
				</div>
			</div>
		{/if}
	</div>

	<!-- Mini story if provided -->
	{#if miniStory}
		<p class="mt-2 text-sm leading-relaxed text-muted-foreground">
			{miniStory}
		</p>
	{/if}
</a>
