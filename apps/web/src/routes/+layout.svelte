<script lang="ts">
	import './layout.css';
	import favicon from '#lib/assets/favicon.png';
	import AlertCircle from '@lucide/svelte/icons/circle-alert';
	import { browser } from '$app/env';
	import { beforeNavigate, afterNavigate, onNavigate } from '$app/navigation';
	import posthog from 'posthog-js';
	import { ModeWatcher } from 'mode-watcher';
	import Navbar from '#lib/components/Navbar.svelte';
	import type { LayoutData } from './$types';
	import type { Snippet } from 'svelte';

	let { children, data }: { children: Snippet; data: LayoutData } = $props();

	if (browser) {
		posthog.init('phc_jeqQwskahiaP3wSMZH9WKRkS3Sj6373SAw4pJAWja9m', {
			api_host: '/ingest',
			ui_host: 'https://us.i.posthog.com',
			person_profiles: 'identified_only',
			capture_pageview: false,
			capture_pageleave: false
		});

		beforeNavigate(({ shallow, type }) => {
			if (shallow && type === 'goto') return;

			return posthog.capture('$pageleave');
		});

		afterNavigate(({ shallow, type }) => {
			if (shallow && type === 'goto') return;

			return posthog.capture('$pageview');
		});
	}

	onNavigate((navigation) => {
		if (navigation.shallow && navigation.type === 'goto') return;
		if (!document.startViewTransition) return;

		return new Promise((resolve) => {
			document.startViewTransition(async () => {
				resolve();
				await navigation.complete;
			});
		});
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<script type="application/ld+json">
		{
			"@context": "https://schema.org",
			"@type": "WebSite",
			"name": "NTX",
			"url": "https://ntx.anishshrestha.com",
			"description": "Real-time Nepal Stock Exchange (NEPSE) explorer and analytics platform.",
			"potentialAction": {
				"@type": "SearchAction",
				"target": "https://ntx.anishshrestha.com/company?q={search_term_string}",
				"query-input": "required name=search_term_string"
			}
		}
	</script>
	<script type="application/ld+json">
		{
			"@context": "https://schema.org",
			"@type": "Organization",
			"name": "NTX",
			"url": "https://ntx.anishshrestha.com",
			"logo": "https://ntx.anishshrestha.com/favicon.png",
			"sameAs": ["https://github.com/voidarchive/ntx"]
		}
	</script>
</svelte:head>
<ModeWatcher />
<Navbar companies={data.companies ?? []} prices={data.prices ?? []} />
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
					<svg class="size-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"
						><path
							d="M12 .3a12 12 0 0 0-3.8 23.4c.6.1.8-.3.8-.6v-2c-3.3.7-4-1.6-4-1.6-.6-1.4-1.4-1.8-1.4-1.8-1-.7.1-.7.1-.7 1.2.1 1.8 1.2 1.8 1.2 1 1.8 2.8 1.3 3.5 1 .1-.8.4-1.3.7-1.6-2.7-.3-5.5-1.3-5.5-5.9 0-1.3.5-2.4 1.2-3.2-.1-.3-.5-1.5.1-3.2 0 0 1-.3 3.3 1.2a11.5 11.5 0 0 1 6 0C17.3 4.7 18.3 5 18.3 5c.6 1.7.2 2.9.1 3.2.8.8 1.2 1.9 1.2 3.2 0 4.6-2.8 5.6-5.5 5.9.4.4.8 1.1.8 2.2v3.3c0 .3.2.7.8.6A12 12 0 0 0 12 .3z"
						/></svg
					>
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
