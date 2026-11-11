import { error } from '@sveltejs/kit';
import { apiBaseUrl, createApiClient } from '#lib/api/client.js';
import { OWNER_COOKIE } from '#lib/owner.js';

// The personal file as a download; the API writes it, this only names it.
export const GET = async ({ cookies }) => {
	const token = cookies.get(OWNER_COOKIE);
	if (!token) throw error(401, 'Enter the owner token first.');
	const res = await createApiClient(apiBaseUrl(), token).portfolio.exportPersonal({});
	const day = new Date().toISOString().slice(0, 10);
	return new Response(new TextDecoder().decode(res.json), {
		headers: {
			'Content-Type': 'application/json',
			'Content-Disposition': `attachment; filename="ntx-personal-${day}.json"`
		}
	});
};
