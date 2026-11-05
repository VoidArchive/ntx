<script lang="ts">
	import Menu from '@lucide/svelte/icons/menu';
	import { page } from '$app/state';
	import { SvelteSet } from 'svelte/reactivity';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import SearchCommand from './SearchCommand.svelte';
	import ThemeToggle from './ThemeToggle.svelte';
	import type { SearchEntry } from './SearchCommand.svelte';

	interface Props {
		companies?: SearchEntry[];
	}

	let { companies = [] }: Props = $props();

	const links = [
		{ href: '/', label: 'Home' },
		{ href: '/company', label: 'Companies' },
		{ href: '/mutual-funds', label: 'Mutual funds' }
	];

	// Company page sections, in page order.
	const sections = [
		{ id: 'earnings', label: 'Earnings' },
		{ id: 'valuation', label: 'Valuation' },
		{ id: 'payout', label: 'Payout' },
		{ id: 'ownership', label: 'Ownership' },
		{ id: 'funds', label: 'Funds' },
		{ id: 'price', label: 'Price and volume' }
	];

	let pathname = $derived(page.url.pathname);
	let onCompanyPage = $derived(/^\/company\/[^/]+$/.test(pathname));
	let activeSection = $state('');
	let menuOpen = $state(false);

	function isCurrent(href: string): boolean {
		if (href === '/') return pathname === '/';
		return pathname.startsWith(href);
	}

	// The topmost visible section wins; a long section stays active while scrolled through.
	$effect(() => {
		if (!onCompanyPage) return;
		const visible = new SvelteSet<string>();
		const observer = new IntersectionObserver(
			(entries) => {
				for (const e of entries) {
					if (e.isIntersecting) visible.add(e.target.id);
					if (!e.isIntersecting) visible.delete(e.target.id);
				}
				activeSection = sections.find((s) => visible.has(s.id))?.id ?? activeSection;
			},
			{ rootMargin: '-10% 0px -60% 0px' }
		);
		for (const s of sections) {
			const el = document.getElementById(s.id);
			if (el) observer.observe(el);
		}
		return () => observer.disconnect();
	});
</script>

{#snippet nav()}
	<ul class="space-y-0.5 text-sm">
		{#each links as link (link.href)}
			<li>
				<a
					href={link.href}
					aria-current={isCurrent(link.href) ? 'page' : undefined}
					onclick={() => (menuOpen = false)}
					class="block rounded px-2 py-1.5 transition-colors hover:bg-sidebar-accent {isCurrent(
						link.href
					)
						? 'font-medium text-sidebar-foreground'
						: 'text-muted-foreground'}"
				>
					{link.label}
				</a>
			</li>
		{/each}
	</ul>

	{#if onCompanyPage}
		<p class="mt-6 mb-1 px-2 text-xs text-muted-foreground">On this page</p>
		<ul class="space-y-0.5 border-l border-sidebar-border text-sm">
			{#each sections as s (s.id)}
				<li>
					<a
						href="#{s.id}"
						onclick={() => (menuOpen = false)}
						class="-ml-px block border-l py-1 pl-3 transition-colors {activeSection === s.id
							? 'border-foreground text-sidebar-foreground'
							: 'border-transparent text-muted-foreground hover:text-foreground'}"
					>
						{s.label}
					</a>
				</li>
			{/each}
		</ul>
	{/if}
{/snippet}

<!-- Desktop: one column pinned to the viewport -->
<aside
	class="sticky top-0 hidden h-screen w-56 shrink-0 flex-col border-r border-sidebar-border bg-sidebar px-3 py-4 lg:flex"
>
	<a href="/" class="px-2 font-serif text-xl tracking-tight hover:opacity-80">NTX</a>
	{#if companies.length > 0}
		<div class="mt-4">
			<SearchCommand {companies} variant="compact" placeholder="Search stocks..." />
		</div>
	{/if}
	<nav class="mt-4 flex-1 overflow-y-auto" aria-label="Site">
		{@render nav()}
	</nav>
	<div class="flex justify-end">
		<ThemeToggle />
	</div>
</aside>

<!-- Phone: slim bar, menu opens the same list -->
<header
	class="sticky top-0 z-40 flex h-12 items-center justify-between border-b border-border bg-background/80 px-4 backdrop-blur-sm lg:hidden"
>
	<a href="/" class="font-serif text-lg tracking-tight">NTX</a>
	<Dialog.Root bind:open={menuOpen}>
		<Dialog.Trigger class="rounded p-1.5 hover:bg-muted" aria-label="Open menu">
			<Menu class="size-5" />
		</Dialog.Trigger>
		<Dialog.Content class="top-0 left-0 h-full max-w-xs translate-x-0 translate-y-0 content-start">
			<Dialog.Title class="font-serif text-lg">NTX</Dialog.Title>
			{#if companies.length > 0}
				<SearchCommand {companies} variant="compact" placeholder="Search stocks..." />
			{/if}
			<nav aria-label="Site">
				{@render nav()}
			</nav>
			<div class="flex justify-end">
				<ThemeToggle />
			</div>
		</Dialog.Content>
	</Dialog.Root>
</header>
