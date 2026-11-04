<script lang="ts">
	import type { OwnershipObservation } from '#lib/gen/ntx/v1/story_pb.js';

	interface Props {
		observations: OwnershipObservation[];
	}

	let { observations }: Props = $props();

	let newestFirst = $derived([...observations].reverse());

	function pct(value: number | undefined): string {
		return value === undefined ? '—' : `${value.toFixed(2)}%`;
	}
</script>

<section id="ownership">
	<h3 class="mb-4 font-serif text-base font-medium">Ownership over time</h3>
	{#if newestFirst.length === 0}
		<p class="text-sm text-muted-foreground">No shareholding observations recorded.</p>
	{:else}
		<ul class="space-y-3">
			{#each newestFirst as o (o.asOf)}
				<li>
					<div class="flex justify-between text-xs text-muted-foreground tabular-nums">
						<span>Promoter {pct(o.promoter)}</span>
						<span>{o.asOf}</span>
						<span>Public {pct(o.public)}</span>
					</div>
					<!-- Labels above carry the values; the bar only shows the split. -->
					<div class="mt-1 flex h-2 overflow-hidden rounded-full bg-muted" aria-hidden="true">
						<div class="bg-foreground" style="width: {o.promoter ?? 0}%"></div>
						<div
							class="border-l-2 border-background bg-muted-foreground/40"
							style="width: {o.public ?? 0}%"
						></div>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</section>
