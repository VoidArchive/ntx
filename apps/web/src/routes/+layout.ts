import posthog from 'posthog-js';
import { browser } from '$app/environment';

export const load = async () => {
	if (browser) {
		posthog.init('phc_jeqQwskahiaP3wSMZH9WKRkS3Sj6373SAw4pJAWja9m', {
			api_host: '/ingest',
			ui_host: 'https://us.i.posthog.com',
			person_profiles: 'identified_only'
		});
	}

	return;
};
