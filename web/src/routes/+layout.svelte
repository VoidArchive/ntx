<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { Header, Footer, MobileNav } from '$lib/components/layout';
	import { SearchDialog } from '$lib/components/search';

	let { children } = $props();

	let searchOpen = $state(false);

	// Initialize theme from localStorage
	$effect(() => {
		if (typeof window !== 'undefined') {
			const stored = localStorage.getItem('theme');
			const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
			const isDark = stored ? stored === 'dark' : prefersDark;
			document.documentElement.classList.toggle('dark', isDark);
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>NTX - NEPSE Stock Data Aggregator</title>
	<meta
		name="description"
		content="Open-source NEPSE data aggregator with screening capabilities. Access stock data, company fundamentals, and market insights."
	/>
</svelte:head>

<div class="flex min-h-screen flex-col">
	<Header onSearchClick={() => (searchOpen = true)} />

	<main class="flex-1 pb-16 md:pb-0">
		{@render children()}
	</main>

	<Footer />
	<MobileNav />
</div>

<SearchDialog bind:open={searchOpen} />
