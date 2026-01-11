<script lang="ts">
	import type { Company, Price, Fundamental } from '$lib/gen/ntx/v1/common_pb';
	import type { SectorStats } from '$lib/gen/ntx/v1/company_pb';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import Sparkles from '@lucide/svelte/icons/sparkles';
	import Copy from '@lucide/svelte/icons/copy';
	import Check from '@lucide/svelte/icons/check';
	import { Sector } from '$lib/gen/ntx/v1/common_pb';

	interface Props {
		company?: Company;
		price?: Price;
		fundamentals?: Fundamental;
		sectorStats?: SectorStats;
	}

	let { company, price, fundamentals, sectorStats }: Props = $props();

	let dialogOpen = $state(false);
	let copied = $state(false);

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

	function fmt(value: number | bigint | undefined): string {
		if (value === undefined) return '—';
		return Number(value).toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	let currentPrice = $derived(price?.ltp ?? price?.close);

	let prompt = $derived.by(() => {
		if (!company || !price) return '';

		const sector = sectorNames[company.sector ?? Sector.OTHERS];
		const today = price.businessDate ?? new Date().toISOString().split('T')[0];

		return `Act as a Senior Financial Analyst specializing in the Nepalese Stock Market (NEPSE).
Your goal is to perform a deep-dive investment analysis of: ${company.name} (${company.symbol}).

## 1. Provided Data Snapshot (As of ${today})
- **Price**: Rs. ${fmt(currentPrice)}
- **Sector**: ${sector}
- **Sector Avg P/E**: ${fmt(sectorStats?.avgPeRatio ?? 0)}
- **Fundamentals**:
  - EPS: ${fmt(fundamentals?.eps)}
  - P/E Ratio: ${fmt(fundamentals?.peRatio)}
  - Book Value: ${fmt(fundamentals?.bookValue)}
  - Paid-up Capital: ${fmt(fundamentals?.paidUpCapital)}

## 2. Research Tasks (MANDATORY WEB SEARCH)
Please SEARCH THE WEB (using browsing capabilities) for the following real-time information:
1.  **Recent News**: Look for the latest news on "sharesansar", "merolagani", or "bizmandu" regarding ${company.symbol} in the last 6 months.
2.  **Regulatory Impacts**: Are there any recent NRB directives, BFI regulations, or insurance board policies affecting the ${sector} sector?
3.  **Corporate Actions**: Check for recent AGM announcements, dividend declarations, or right share issues.

## 3. Analysis Requirements
Combine the provided data with your web research to answer:
- **Valuation**: Is ${company.symbol} undervalued compared to its peers in the ${sector} sector? (Compare P/E and P/B).
- **Growth Outlook**: Based on the latest quarterly reports you find, is the company growing its core business?
- **Risk Assessment**: What are the specific regulatory or macro risks for this company right now?

## 4. Investment Verdict
Conclude with a structured verdict:
- **Recommendation**: [Buy / Hold / Sell]
- **Time Horizon**: [Short-term / Long-term]
- **Key Catalyst**: [One specific event to watch]

Please be objective, critical, and data-driven.`;
	});

	async function copyPrompt() {
		if (!prompt) return;
		await navigator.clipboard.writeText(prompt);
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}
</script>

<Button variant="outline" size="sm" onclick={() => (dialogOpen = true)} class="w-full">
	<Sparkles class="size-4" />
	AI Research
</Button>

<Dialog.Root bind:open={dialogOpen}>
	<Dialog.Content class="max-w-2xl">
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<Sparkles class="size-5" />
				AI Research Prompt
			</Dialog.Title>
			<Dialog.Description>
				Copy this prompt and paste it into your preferred AI assistant (ChatGPT, Claude, etc.) for
				detailed investment research.
			</Dialog.Description>
		</Dialog.Header>
		<div class="relative">
			<pre
				class="max-h-[400px] overflow-auto rounded-lg bg-muted p-4 text-sm whitespace-pre-wrap">{prompt}</pre>
			<Button variant="secondary" size="sm" class="absolute top-2 right-2" onclick={copyPrompt}>
				{#if copied}
					<Check class="size-4" />
					Copied!
				{:else}
					<Copy class="size-4" />
					Copy
				{/if}
			</Button>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (dialogOpen = false)}>Close</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
