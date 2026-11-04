import { defineEnvVars } from '@sveltejs/kit/env';

// PostHog is optional: identity validator lets the value be undefined.
const optional = (value: string | undefined) => value;

// The Worker var from wrangler.jsonc (adapter-cloudflare 8 dropped event.platform). Optional here
// because the build has no bindings; apiBaseUrl() refuses to run without it in production.
export const variables = defineEnvVars({
	API_URL: { schema: optional, description: 'Base URL of the Go API' },
	PUBLIC_POSTHOG_KEY: { public: true, static: true, schema: optional },
	PUBLIC_POSTHOG_HOST: { public: true, static: true, schema: optional }
});
