<script lang="ts">
	import { goto } from '$app/navigation';
	import { createApiClient } from '$lib/api/client';
	import { authStore } from '$lib/stores/auth.svelte';
	import Lock from '@lucide/svelte/icons/lock';
	import Mail from '@lucide/svelte/icons/mail';
	import Loader2 from '@lucide/svelte/icons/loader-2';

	const API_URL = import.meta.env.DEV ? 'http://localhost:8080' : 'https://ntx-api.anishshrestha.com';

	let email = $state('');
	let password = $state('');
	let error = $state<string | null>(null);
	let isLoading = $state(false);

	// If already authenticated, redirect
	$effect(() => {
		if (authStore.state.isAuthenticated) {
			goto('/portfolio');
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = null;
		isLoading = true;

		try {
			const api = createApiClient(API_URL);
			const response = await api.auth.login({ email, password });
			authStore.login(response.token, response.userId);
			goto('/portfolio');
		} catch (err) {
			error = 'Invalid email or password';
			console.error('Login failed:', err);
		} finally {
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Login - NTX Portfolio</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<div class="flex min-h-[80vh] items-center justify-center px-4">
	<div class="w-full max-w-md">
		<div class="rounded-2xl border border-border bg-card/50 p-8 shadow-xl backdrop-blur-sm">
			<div class="mb-8 text-center">
				<h1 class="font-serif text-2xl font-medium">Welcome back</h1>
				<p class="mt-2 text-sm text-muted-foreground">Sign in to your portfolio</p>
			</div>

			{#if error}
				<div class="mb-6 rounded-lg border border-destructive/50 bg-destructive/10 px-4 py-3 text-sm text-destructive">
					{error}
				</div>
			{/if}

			<form onsubmit={handleSubmit} class="space-y-5">
				<div>
					<label for="email" class="mb-2 block text-sm font-medium">Email</label>
					<div class="relative">
						<Mail class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
						<input
							id="email"
							type="email"
							bind:value={email}
							required
							autocomplete="email"
							class="w-full rounded-lg border border-border bg-background py-2.5 pl-10 pr-4 text-sm transition-colors focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
							placeholder="you@example.com"
						/>
					</div>
				</div>

				<div>
					<label for="password" class="mb-2 block text-sm font-medium">Password</label>
					<div class="relative">
						<Lock class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
						<input
							id="password"
							type="password"
							bind:value={password}
							required
							autocomplete="current-password"
							class="w-full rounded-lg border border-border bg-background py-2.5 pl-10 pr-4 text-sm transition-colors focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
							placeholder="••••••••"
						/>
					</div>
				</div>

				<button
					type="submit"
					disabled={isLoading}
					class="flex w-full items-center justify-center gap-2 rounded-lg bg-primary py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
				>
					{#if isLoading}
						<Loader2 class="size-4 animate-spin" />
						Signing in...
					{:else}
						Sign in
					{/if}
				</button>
			</form>

			<p class="mt-6 text-center text-sm text-muted-foreground">
				Don't have an account?
				<a href="/register" class="font-medium text-primary hover:underline">Create one</a>
			</p>
		</div>
	</div>
</div>
