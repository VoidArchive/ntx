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

		return `You are helping research ${company.name} (${company.symbol}), listed on the Nepal Stock Exchange (NEPSE).
Check the claims and checks below against the company's own filings; do not give advice or a recommendation.

## 1. Data from NTX (close of ${quote?.businessDate ?? 'no recorded close'})
- **Price**: Rs. ${fmt(quote?.close)}
- **Sector**: ${sectorNames[company.sector] ?? 'Others'}
- **Newest filing**: ${story?.lastFiling || 'none'}

What the filings say:
${claims || '- Not enough filings to say anything yet.'}

Ratios against the sector:
${strips || '- None recorded.'}

Published checks:
${checks || '- None.'}

## 2. Check against the filings
1. For each claim and check above, find the quarterly report, annual report or notice it rests on (nepalstock.com.np, the company's website) and say whether the figure matches.
2. Note any filing, AGM notice, dividend or rights announcement newer than ${story?.lastFiling || 'the newest filing above'}.
3. Note any NRB directive or regulator policy in the last six months that applies to this sector.

## 3. What would change them
For each claim and check, list the specific figure or event in a coming filing that would make it no longer hold.

Cite the filing or page for every figure you use, and say "not found" where you cannot find one.`;
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
				Copy this prompt and paste it into your preferred AI assistant (ChatGPT, Claude, etc.) to
				check what the filings say.
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
