import { defineEnvVars } from '@sveltejs/kit/env';

// PostHog is optional: identity validator lets the value be undefined.
const optional = (value: string | undefined) => value;

export const variables = defineEnvVars({
	PUBLIC_POSTHOG_KEY: { public: true, static: true, schema: optional },
	PUBLIC_POSTHOG_HOST: { public: true, static: true, schema: optional }
});
