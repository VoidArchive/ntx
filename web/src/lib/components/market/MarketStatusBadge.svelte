<script lang="ts">
	let {
		isOpen,
		state
	}: {
		isOpen: boolean;
		state: string;
	} = $props();

	const stateLabels: Record<string, string> = {
		open: 'Market Open',
		closed: 'Market Closed',
		'pre-open': 'Pre-Open',
		'pre-close': 'Pre-Close'
	};
</script>

<div
	class="inline-flex items-center gap-2 rounded-full px-3 py-1 text-sm font-medium {isOpen
		? 'bg-positive-muted text-positive'
		: 'bg-muted text-muted-foreground'}"
>
	<span class="relative flex h-2 w-2">
		{#if isOpen}
			<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-positive opacity-75"></span>
		{/if}
		<span
			class="relative inline-flex h-2 w-2 rounded-full {isOpen ? 'bg-positive' : 'bg-muted-foreground'}"
		></span>
	</span>
	{stateLabels[state] ?? state}
</div>
