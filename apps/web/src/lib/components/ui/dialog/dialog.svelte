<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import type { Snippet } from 'svelte';

	interface Props extends Omit<DialogPrimitive.RootProps, 'children'> {
		open?: boolean;
		children?: Snippet;
	}

	let {
		open = $bindable(false),
		children,
		...restProps
	}: Props = $props();
</script>

<DialogPrimitive.Root bind:open {...restProps}>
	{@render children?.()}
</DialogPrimitive.Root>
