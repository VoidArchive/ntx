<script lang="ts">
	import { formatVolume } from '$lib/utils/format';
	import { getSectorName, getSectorColor } from '$lib/utils/sector';
	import { Sector } from '@ntx/api/ntx/v1/common_pb';
	import { Building2Icon, ArrowRightIcon } from '@lucide/svelte';

	let {
		sector,
		stockCount,
		turnover,
		href
	}: {
		sector: Sector;
		stockCount: number;
		turnover: bigint | number;
		href?: string;
	} = $props();
</script>

<a
	href={href ?? `/companies?sector=${sector}`}
	class="group flex flex-col rounded-xl border bg-card p-4 transition-colors hover:bg-accent/50"
>
	<div class="flex items-start justify-between">
		<div class="rounded-lg p-2 {getSectorColor(sector)}">
			<Building2Icon class="h-5 w-5" />
		</div>
		<ArrowRightIcon
			class="h-4 w-4 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
		/>
	</div>
	<div class="mt-3">
		<div class="font-medium">{getSectorName(sector)}</div>
		<div class="mt-1 flex items-center gap-3 text-xs text-muted-foreground">
			<span>{stockCount} stocks</span>
			<span>Turnover: {formatVolume(turnover)}</span>
		</div>
	</div>
</a>
