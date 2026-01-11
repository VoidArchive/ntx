import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	if (event.url.pathname.startsWith('/ingest')) {
		const originalUrl = event.url.pathname.replace('/ingest', '');
		const upstreamUrl = `https://us.i.posthog.com${originalUrl}${event.url.search}`;

		return fetch(upstreamUrl, {
			method: event.request.method,
			headers: {
				...Object.fromEntries(event.request.headers),
				host: 'us.i.posthog.com'
			},
			body: event.request.body,
			// @ts-ignore
			duplex: 'half'
		});
	}

	return resolve(event);
};
