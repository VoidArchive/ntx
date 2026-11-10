import type { Cookies } from '@sveltejs/kit';

// The owner token lives in one httpOnly cookie so server loads can forward it;
// it never reaches page scripts (ADR 0006).
export const OWNER_COOKIE = 'ntx_owner';

export function rememberOwner(cookies: Cookies, token: string) {
	cookies.set(OWNER_COOKIE, token, {
		path: '/',
		httpOnly: true,
		secure: !import.meta.env.DEV,
		sameSite: 'lax',
		maxAge: 60 * 60 * 24 * 365
	});
}

export function forgetOwner(cookies: Cookies) {
	cookies.delete(OWNER_COOKIE, { path: '/' });
}
