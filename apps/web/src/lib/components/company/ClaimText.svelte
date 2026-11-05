<script lang="ts">
	import type { Claim } from '#lib/gen/ntx/v1/story_pb.js';

	interface Props {
		claims: Claim[];
	}

	let { claims }: Props = $props();

	// Claim text is computed in the API; its numbers (or, without numbers, the
	// whole sentence) link to their evidence, and its period and filing follow it.
	const NUMBER = /(Rs [\d,]+(?:\.\d+)? crore|Rs [\d,]+(?:\.\d+)?|[\d,]+(?:\.\d+)?(?:%|x))/;
</script>

{#each claims as claim, i (i)}
	{@const parts = claim.text.split(NUMBER)}
	{#each parts.length === 1 ? ['', claim.text, ''] : parts as part, j (j)}
		{#if j % 2 === 1}
			<a
				href="#{claim.anchor}"
				class="underline decoration-dotted underline-offset-4 hover:text-primary">{part}</a
			>
		{:else}{part}{/if}
	{/each}{#if claim.period && !claim.text.includes(claim.period)}
		<span class="text-xs font-normal text-muted-foreground">({claim.period})</span
		>{/if}{#if claim.source}
		<a
			href={claim.source}
			class="text-xs font-normal text-muted-foreground underline underline-offset-2 hover:text-primary"
			>filing</a
		>{/if}{i < claims.length - 1 ? ' ' : ''}
{/each}
