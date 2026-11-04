<script lang="ts">
	import type { Company } from '#lib/gen/ntx/v1/common_pb.js';
	import type { GetCompanyStoryResponse } from '#lib/gen/ntx/v1/story_pb.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import Sparkles from '@lucide/svelte/icons/sparkles';
	import Copy from '@lucide/svelte/icons/copy';
	import Check from '@lucide/svelte/icons/check';
	import { sectorNames } from '#lib/utils.js';

	interface Props {
		company: Company;
		story?: GetCompanyStoryResponse;
	}

	let { company, story }: Props = $props();

	let dialogOpen = $state(false);
	let copied = $state(false);

	function fmt(value: number | undefined): string {
		if (value === undefined) return '—';
		return value.toLocaleString('en-NP', { maximumFractionDigits: 2 });
	}

	// The claims go in verbatim so the assistant starts from what the filings say.
	let prompt = $derived.by(() => {
		const quote = story?.quote;
		const claims = (story?.claims ?? []).map((c) => `- ${c.text}`).join('\n');
		const strips = (story?.strips ?? [])
			.map(
				(s) =>
					`- ${s.ratio.toUpperCase()}: ${fmt(s.value)} (sector median ${fmt(s.sector?.median)})`
			)
			.join('\n');
		const checks = (story?.checks ?? [])
			.map((c) => `- ${c.name}: ${c.verdict} (${c.rule})`)
			.join('\n');

		return `Act as a Senior Financial Analyst specializing in the Nepalese Stock Market (NEPSE).
Your goal is to perform a deep-dive investment analysis of: ${company.name} (${company.symbol}).

## 1. Provided Data Snapshot (As of ${quote?.businessDate ?? 'no recorded close'})
- **Price**: Rs. ${fmt(quote?.close)}
- **Sector**: ${sectorNames[company.sector] ?? 'Others'}
- **Newest filing**: ${story?.lastFiling || 'none'}

What the filings say:
${claims || '- Not enough filings to say anything yet.'}

Ratios against the sector:
${strips || '- None recorded.'}

Published checks:
${checks || '- None.'}

## 2. Research Tasks (MANDATORY WEB SEARCH)
Please SEARCH THE WEB (using browsing capabilities) for the following real-time information:
1.  **Recent News**: Look for the latest news on "sharesansar", "merolagani", or "bizmandu" regarding ${company.symbol} in the last 6 months.
2.  **Regulatory Impacts**: Are there any recent NRB directives, BFI regulations, or insurance board policies affecting this sector?
3.  **Corporate Actions**: Check for recent AGM announcements, dividend declarations, or right share issues.

## 3. Analysis Requirements
Combine the provided data with your web research to answer:
- **Valuation**: Is ${company.symbol} undervalued compared to its sector peers? (Compare P/E and P/B).
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

<Button variant="outline" size="sm" onclick={() => (dialogOpen = true)}>
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
