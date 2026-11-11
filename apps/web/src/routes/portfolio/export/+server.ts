import { Code, ConnectError } from '@connectrpc/connect';
import { error, redirect } from '@sveltejs/kit';
import { apiBaseUrl, createApiClient } from '#lib/api/client.js';
import { OWNER_COOKIE, forgetOwner } from '#lib/owner.js';

// The personal file as a download; the API writes it, this only names it.
export const GET = async ({ cookies }) => {
	const token = cookies.get(OWNER_COOKIE);
	if (!token) throw error(401, 'Enter the owner token first.');
	let file: Uint8Array;
	try {
		file = (await createApiClient(apiBaseUrl(), token).portfolio.exportPersonal({})).json;
	} catch (e) {
		// A stale token reads the same as none: back to the token field.
		if (e instanceof ConnectError && e.code === Code.Unauthenticated) {
			forgetOwner(cookies);
			throw redirect(303, '/portfolio');
		}
		throw e;
	}
	const day = new Date().toISOString().slice(0, 10);
	return new Response(new TextDecoder().decode(file), {
		headers: {
			'Content-Type': 'application/json',
			'Content-Disposition': `attachment; filename="ntx-personal-${day}.json"`,
			'Cache-Control': 'no-store'
		}
	});
};
