<script lang="ts">
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';
	import { buttonVariants, type Variant, type Size } from './index';

	interface Props extends HTMLButtonAttributes {
		variant?: Variant;
		size?: Size;
		children?: Snippet;
	}

	let {
		class: className,
		variant = 'default',
		size = 'default',
		children,
		...restProps
	}: Props = $props();
</script>

<button class={cn(buttonVariants({ variant, size }), className)} {...restProps}>
	{@render children?.()}
</button>
