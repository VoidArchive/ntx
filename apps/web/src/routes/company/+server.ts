import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// The list moved to /screener; old links and saved filter URLs keep working.
export const GET: RequestHandler = ({ url }) => {
	redirect(301, `/screener${url.search}`);
};
