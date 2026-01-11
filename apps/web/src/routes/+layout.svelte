<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.png';
	import Github from '@lucide/svelte/icons/github';
	import AlertCircle from '@lucide/svelte/icons/alert-circle';
	import { browser } from '$app/environment';
	import { beforeNavigate, afterNavigate } from '$app/navigation';
	import posthog from 'posthog-js';
	import { ModeWatcher } from 'mode-watcher';

	let { children } = $props();

	if (browser) {
		beforeNavigate(() => posthog.capture('$pageleave'));
		afterNavigate(() => posthog.capture('$pageview'));
	}
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
<ModeWatcher />
{@render children()}

<footer class="mt-20 border-t border-border/50 bg-background/50 py-12 backdrop-blur-sm">
	<div class="mx-auto max-w-7xl px-4 text-center">
		<div class="mb-6 flex justify-center gap-6">
			<a
				href="https://github.com/voidarchive/ntx"
				target="_blank"
				rel="noreferrer"
				class="group flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
			>
				<div
					class="flex size-8 items-center justify-center rounded-full bg-muted/50 transition-colors group-hover:bg-primary/10"
				>
					<Github class="size-4" />
				</div>
				<span>View on GitHub</span>
			</a>

			<a
				href="https://github.com/voidarchive/ntx/issues"
				target="_blank"
				rel="noreferrer"
				class="group flex items-center gap-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
			>
				<div
					class="flex size-8 items-center justify-center rounded-full bg-muted/50 transition-colors group-hover:bg-orange-500/10 group-hover:text-orange-600"
				>
					<AlertCircle class="size-4" />
				</div>
				<span>Report Issue</span>
			</a>
		</div>

		<p class="text-xs text-muted-foreground/60">
			Built by <a
				target="_blank"
				href="https://anishshrestha.com"
				class="hover:text-foreground hover:underline">Anish Shrestha</a
			>.
		</p>
	</div>
</footer>
