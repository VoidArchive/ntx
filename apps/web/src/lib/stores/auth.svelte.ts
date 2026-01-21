// Auth store for managing authentication state
import { browser } from '$app/environment';

const TOKEN_KEY = 'ntx_auth_token';
const USER_ID_KEY = 'ntx_user_id';

interface AuthState {
	token: string | null;
	userId: bigint | null;
	isAuthenticated: boolean;
}

function createAuthStore() {
	let token = $state<string | null>(null);
	let userId = $state<bigint | null>(null);

	// Load from localStorage on init
	if (browser) {
		const savedToken = localStorage.getItem(TOKEN_KEY);
		const savedUserId = localStorage.getItem(USER_ID_KEY);
		if (savedToken) {
			token = savedToken;
			userId = savedUserId ? BigInt(savedUserId) : null;
		}
	}

	function login(newToken: string, newUserId: bigint) {
		token = newToken;
		userId = newUserId;
		if (browser) {
			localStorage.setItem(TOKEN_KEY, newToken);
			localStorage.setItem(USER_ID_KEY, newUserId.toString());
		}
	}

	function logout() {
		token = null;
		userId = null;
		if (browser) {
			localStorage.removeItem(TOKEN_KEY);
			localStorage.removeItem(USER_ID_KEY);
		}
	}

	function getToken(): string | null {
		return token;
	}

	return {
		get state(): AuthState {
			return {
				token,
				userId,
				isAuthenticated: !!token
			};
		},
		login,
		logout,
		getToken
	};
}

export const authStore = createAuthStore();
