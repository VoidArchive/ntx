<script lang="ts">
	import Menu from '@lucide/svelte/icons/menu';
	import House from '@lucide/svelte/icons/house';
	import SlidersHorizontal from '@lucide/svelte/icons/sliders-horizontal';
	import Droplets from '@lucide/svelte/icons/droplets';
	import ChartPie from '@lucide/svelte/icons/chart-pie';
	import { page } from '$app/state';
	import { SvelteSet } from 'svelte/reactivity';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import SearchCommand from './SearchCommand.svelte';
	import ThemeToggle from './ThemeToggle.svelte';
	import type { SearchEntry } from './SearchCommand.svelte';

	interface Props {
		companies?: SearchEntry[];
		owner?: boolean;
	}

	let { companies = [], owner = false }: Props = $props();

	// Portfolio is the owner's page; without the token it is not offered (ADR 0006).
	// The first four are the phone's tabs; short is the label under the tab icon.
	const tabs = [
		{ href: '/', label: 'Home', short: 'Home', icon: House },
		{ href: '/screener', label: 'Screener', short: 'Screener', icon: SlidersHorizontal },
		{ href: '/hydro', label: 'Hydro', short: 'Hydro', icon: Droplets },
		{ href: '/mutual-funds', label: 'Mutual funds', short: 'Funds', icon: ChartPie }
	];
	let links = $derived([...tabs, ...(owner ? [{ href: '/portfolio', label: 'Portfolio' }] : [])]);

	// Saved views of the company list; each is a URL the screener already understands.
	const lists = [
		{ href: '/screener?sort=change&dir=desc', label: 'Gainers' },
		{ href: '/screener?sort=change&dir=asc', label: 'Losers' },
		{ href: '/screener?sort=turnover&dir=desc', label: 'Most traded' },
		{ href: '/screener?sort=market_cap&dir=desc', label: 'Largest market cap' }
	];

	// Company page sections, in the report's reading order; the number is the order the reader follows.
	const sections = [
		{ id: 'overview', label: 'Overview' },
		{ id: 'valuation', n: 1, label: 'Valuation' },
		{ id: 'earnings', n: 2, label: 'Past Performance' },
		{ id: 'health', n: 3, label: 'Financial Health' },
		{ id: 'payout', n: 4, label: 'Dividend' },
		{ id: 'ownership', n: 5, label: 'Ownership' },
		{ id: 'funds', n: 6, label: 'Funds' },
		{ id: 'plants', n: 7, label: 'Plants' } // hydro companies only
	];

	// A shallow goto leaves page.url on the loaded URL; the target is in page.shallow.
	let current = $derived(page.shallow?.url ?? page.url);
	let pathname = $derived(current.pathname);
	let onCompanyPage = $derived(/^\/company\/[^/]+$/.test(pathname));
	let activeSection = $state('');
	// Sections the current page renders; a company without plants has no Plants link.
	let present = $state<string[]>([]);
	let menuOpen = $state(false);

	function isCurrent(href: string): boolean {
		if (href === '/') return pathname === '/';
		return pathname.startsWith(href);
	}

	// A list link is current only when the screener shows exactly that sort.
	function isList(href: string): boolean {
		const want = new URL(href, 'http://x');
		return (
			pathname === want.pathname &&
			[...want.searchParams].every(([k, v]) => current.searchParams.get(k) === v)
		);
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
		const found: string[] = [];
		for (const s of sections) {
			const el = document.getElementById(s.id);
			if (el) observer.observe(el);
			if (el) found.push(s.id);
		}
		present = found;
		return () => observer.disconnect();
	});
</script>

{#snippet nav(items: typeof links)}
	{#if items.length > 0}
		<ul class="space-y-0.5 text-sm">
			{#each items as link (link.href)}
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
	{/if}

	{#if !onCompanyPage}
		<p class="mt-6 mb-1 px-2 text-xs text-muted-foreground">Lists</p>
		<ul class="space-y-0.5 text-sm" aria-label="Lists">
			{#each lists as list (list.href)}
				<li>
					<a
						href={list.href}
						aria-current={isList(list.href) ? 'page' : undefined}
						onclick={() => (menuOpen = false)}
						class="block rounded px-2 py-1.5 transition-colors hover:bg-sidebar-accent {isList(
							list.href
						)
							? 'font-medium text-sidebar-foreground'
							: 'text-muted-foreground'}"
					>
						{list.label}
					</a>
				</li>
			{/each}
		</ul>
	{/if}

	{#if onCompanyPage}
		<ul class="mt-6 space-y-0.5 text-sm" aria-label="On this page">
			{#each sections.filter((s) => present.length === 0 || present.includes(s.id)) as s (s.id)}
				<li>
					<a
						href="#{s.id}"
						onclick={() => (menuOpen = false)}
						class="flex items-baseline gap-2 rounded-md border-l-2 py-1.5 pr-2 pl-3 transition-colors {activeSection ===
						s.id
							? 'border-primary bg-sidebar-accent font-medium text-sidebar-foreground'
							: 'border-transparent text-muted-foreground hover:text-foreground'}"
					>
						{#if s.n}<span class="w-3 text-xs text-muted-foreground">{s.n}</span>{/if}
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
		{@render nav(links)}
	</nav>
	<div class="flex justify-end">
		<ThemeToggle />
	</div>
</aside>

<!-- Phone: wordmark and search on top, tabs along the bottom, the long lists in a sheet -->
<header
	class="sticky top-0 z-40 flex h-12 items-center gap-3 border-b border-border bg-background/80 px-4 backdrop-blur-sm lg:hidden"
>
	<a href="/" class="font-serif text-lg tracking-tight">NTX</a>
	{#if companies.length > 0}
		<div class="ml-auto w-full max-w-xs">
			<SearchCommand {companies} variant="compact" placeholder="Search stocks..." />
		</div>
	{/if}
</header>
<nav
	aria-label="Site"
	class="fixed inset-x-0 bottom-0 z-40 border-t border-sidebar-border bg-sidebar pb-[env(safe-area-inset-bottom)] lg:hidden"
>
	<ul class="flex h-14 items-stretch">
		{#each tabs as tab (tab.href)}
			<li class="flex-1">
				<a
					href={tab.href}
					aria-current={isCurrent(tab.href) ? 'page' : undefined}
					class="flex h-full flex-col items-center justify-center gap-1 text-[11px] leading-none {isCurrent(
						tab.href
					)
						? 'text-primary'
						: 'text-muted-foreground'}"
				>
					<tab.icon class="size-5" aria-hidden="true" />
					{tab.short}
				</a>
			</li>
		{/each}
		<li class="flex-1">
			<Dialog.Root bind:open={menuOpen}>
				<Dialog.Trigger
					class="flex h-full w-full flex-col items-center justify-center gap-1 text-[11px] leading-none {menuOpen
						? 'text-primary'
						: 'text-muted-foreground'}"
				>
					<Menu class="size-5" aria-hidden="true" />
					Menu
				</Dialog.Trigger>
				<Dialog.Content
					class="top-auto bottom-0 max-h-[85dvh] max-w-none translate-x-[-50%] translate-y-0 content-start overflow-y-auto rounded-t-xl border-b-0 pb-[calc(env(safe-area-inset-bottom)+1.5rem)] data-[state=closed]:slide-out-to-bottom data-[state=open]:slide-in-from-bottom"
				>
					<Dialog.Title class="font-serif text-lg">Menu</Dialog.Title>
					<!-- The tabs are already in the bar; the sheet holds what did not fit. -->
					<nav aria-label="Site">
						{@render nav(links.slice(tabs.length))}
					</nav>
					<div class="flex items-center justify-between border-t border-border pt-4 text-sm">
						<span class="text-muted-foreground">Theme</span>
						<ThemeToggle />
					</div>
				</Dialog.Content>
			</Dialog.Root>
		</li>
	</ul>
</nav>
