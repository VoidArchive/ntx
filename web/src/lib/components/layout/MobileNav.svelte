<script lang="ts">
	import { page } from '$app/stores';
	import { HomeIcon, Building2Icon, PieChartIcon, SlidersHorizontalIcon } from '@lucide/svelte';

	const navItems = [
		{ href: '/', label: 'Home', icon: HomeIcon },
		{ href: '/companies', label: 'Companies', icon: Building2Icon },
		{ href: '/sectors', label: 'Sectors', icon: PieChartIcon },
		{ href: '/screener', label: 'Screener', icon: SlidersHorizontalIcon }
	];

	function isActive(href: string): boolean {
		if (href === '/') return $page.url.pathname === '/';
		return $page.url.pathname.startsWith(href);
	}
</script>

<nav
	class="fixed bottom-0 left-0 right-0 z-50 border-t bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 md:hidden"
>
	<div class="flex justify-around py-2">
		{#each navItems as item}
			<a
				href={item.href}
				class="flex flex-col items-center gap-1 px-3 py-1.5 text-xs transition-colors {isActive(
					item.href
				)
					? 'text-primary'
					: 'text-muted-foreground hover:text-foreground'}"
			>
				<item.icon class="h-5 w-5" />
				<span>{item.label}</span>
			</a>
		{/each}
	</div>
</nav>
