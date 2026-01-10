<script lang="ts">
	import type { Company, Price } from '$lib/gen/ntx/v1/common_pb';
	import { Sector } from '$lib/gen/ntx/v1/common_pb';
	import ArrowRight from '@lucide/svelte/icons/arrow-right';

	interface Props {
		company: Company;
		price?: Price;
	}

	let { company, price }: Props = $props();

	// Use ltp, fallback to close for non-trading days
	let currentPrice = $derived(price?.ltp ?? price?.close);

	function formatPrice(value: number | undefined): string {
		if (value === undefined) return '—';
		return value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	function formatChange(value: number | undefined): string {
		if (value === undefined) return '';
		const prefix = value > 0 ? '+' : '';
		return `${prefix}${value.toFixed(2)}%`;
	}

	// Short labels for pills
	const sectorLabels: Record<number, string> = {
		[Sector.COMMERCIAL_BANK]: 'Bank',
		[Sector.DEVELOPMENT_BANK]: 'Dev Bank',
		[Sector.FINANCE]: 'Finance',
		[Sector.MICROFINANCE]: 'MFI',
		[Sector.LIFE_INSURANCE]: 'Life Ins.',
		[Sector.NON_LIFE_INSURANCE]: 'Non-Life',
		[Sector.HYDROPOWER]: 'Hydro',
		[Sector.MANUFACTURING]: 'Mfg.',
		[Sector.HOTEL]: 'Hotel',
		[Sector.TRADING]: 'Trading',
		[Sector.INVESTMENT]: 'Invest.',
		[Sector.MUTUAL_FUND]: 'MF',
		[Sector.OTHERS]: 'Other'
	};
</script>

<a
	href="/company/{company.symbol}"
	class="group relative flex flex-col justify-between overflow-hidden rounded-xl border border-border bg-card/50 p-6 shadow-sm backdrop-blur-sm transition-all hover:-translate-y-1 hover:border-foreground/20 hover:shadow-md"
>
	<!-- Top: Header -->
	<div class="flex items-start justify-between">
		<div>
			<h3
				class="font-serif text-2xl font-medium tracking-tight text-foreground group-hover:underline"
			>
				{company.symbol}
			</h3>
			<p class="mt-1 line-clamp-2 text-sm text-muted-foreground" title={company.name}>
				{company.name}
			</p>
		</div>

		<!-- Sector Badge (Top Right) -->
		<span
			class="inline-flex items-center rounded-full border border-border bg-background/50 px-2 py-0.5 text-[10px] font-medium tracking-wider text-muted-foreground uppercase"
		>
			{sectorLabels[company.sector ?? Sector.OTHERS] ?? 'Others'}
		</span>
	</div>

	<!-- Middle: Price (Placeholder or Real) -->
	<div class="mt-6 flex items-baseline gap-3">
		{#if currentPrice}
			<span class="text-3xl font-medium text-foreground tabular-nums">
				{formatPrice(currentPrice)}
			</span>
			{#if price.changePercent !== undefined}
				<span
					class="text-sm font-medium tabular-nums {price.changePercent > 0
						? 'text-positive'
						: price.changePercent < 0
							? 'text-negative'
							: 'text-muted-foreground'}"
				>
					{formatChange(price.changePercent)}
				</span>
			{/if}
		{:else}
			<!-- Empty state / Placeholder for visually balanced card -->
			<div class="h-9 w-24 animate-pulse rounded bg-muted/20"></div>
		{/if}
	</div>

	<!-- Bottom: CTA -->
	<div
		class="mt-6 flex items-center justify-between border-t border-border/50 pt-4 opacity-60 transition-opacity group-hover:opacity-100"
	>
		<span class="text-xs font-medium text-muted-foreground">View Analysis</span>
		<ArrowRight
			class="size-4 text-muted-foreground transition-transform group-hover:translate-x-1"
		/>
	</div>

	<!-- Hover Gradient Effect -->
	<div
		class="absolute inset-0 -z-10 bg-gradient-to-br from-primary/5 via-transparent to-transparent opacity-0 transition-opacity group-hover:opacity-100"
	></div>
</a>
