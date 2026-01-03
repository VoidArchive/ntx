<script lang="ts">
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import { SearchIcon, MenuIcon, XIcon, GithubIcon, SunIcon, MoonIcon } from '@lucide/svelte';

	let { onSearchClick = () => {} }: { onSearchClick?: () => void } = $props();

	let menuOpen = $state(false);
	let isDark = $state(false);

	// Check initial theme on mount
	$effect(() => {
		isDark = document.documentElement.classList.contains('dark');
	});

	function toggleTheme() {
		isDark = !isDark;
		document.documentElement.classList.toggle('dark', isDark);
		localStorage.setItem('theme', isDark ? 'dark' : 'light');
	}

	const navItems = [
		{ href: '/screener', label: 'Screener' },
		{ href: '/companies', label: 'Companies' },
		{ href: '/sectors', label: 'Sectors' }
	];

	function isActive(href: string): boolean {
		if (href === '/') return $page.url.pathname === '/';
		return $page.url.pathname.startsWith(href);
	}
</script>

<header
	class="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60"
>
	<div class="mx-auto flex h-14 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
		<!-- Logo -->
		<a href="/" class="flex items-center gap-2">
			<span class="text-xl font-bold tracking-tight">NTX</span>
		</a>

		<!-- Desktop Navigation -->
		<nav class="hidden items-center gap-1 md:flex">
			{#each navItems as item}
				<a
					href={item.href}
					class="rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground {isActive(
						item.href
					)
						? 'bg-accent text-accent-foreground'
						: 'text-muted-foreground'}"
				>
					{item.label}
				</a>
			{/each}
		</nav>

		<!-- Right side actions -->
		<div class="flex items-center gap-2">
			<!-- Search Button (Desktop) -->
			<Button
				variant="outline"
				class="hidden h-9 w-64 justify-start text-sm text-muted-foreground md:flex"
				onclick={onSearchClick}
			>
				<SearchIcon class="mr-2 h-4 w-4" />
				<span>Search companies...</span>
				<kbd
					class="pointer-events-none ml-auto hidden h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-xs font-medium opacity-100 sm:flex"
				>
					<span class="text-xs">⌘</span>K
				</kbd>
			</Button>

			<!-- Search Button (Mobile) -->
			<Button variant="ghost" size="icon" class="md:hidden" onclick={onSearchClick}>
				<SearchIcon class="h-5 w-5" />
				<span class="sr-only">Search</span>
			</Button>

			<!-- Theme Toggle -->
			<Button variant="ghost" size="icon" onclick={toggleTheme}>
				{#if isDark}
					<SunIcon class="h-5 w-5" />
				{:else}
					<MoonIcon class="h-5 w-5" />
				{/if}
				<span class="sr-only">Toggle theme</span>
			</Button>

			<!-- GitHub Link -->
			<Button variant="ghost" size="icon" href="https://github.com/voidarchive/ntx" target="_blank">
				<GithubIcon class="h-5 w-5" />
				<span class="sr-only">GitHub</span>
			</Button>

			<!-- Mobile Menu Toggle -->
			<Button variant="ghost" size="icon" class="md:hidden" onclick={() => (menuOpen = !menuOpen)}>
				{#if menuOpen}
					<XIcon class="h-5 w-5" />
				{:else}
					<MenuIcon class="h-5 w-5" />
				{/if}
				<span class="sr-only">Toggle menu</span>
			</Button>
		</div>
	</div>

	<!-- Mobile Navigation -->
	{#if menuOpen}
		<nav class="border-t px-4 py-3 md:hidden">
			{#each navItems as item}
				<a
					href={item.href}
					onclick={() => (menuOpen = false)}
					class="block rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground {isActive(
						item.href
					)
						? 'bg-accent text-accent-foreground'
						: 'text-muted-foreground'}"
				>
					{item.label}
				</a>
			{/each}
		</nav>
	{/if}
</header>
