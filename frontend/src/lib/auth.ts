import { goto } from '$app/navigation';

const TOKEN_KEY = 'auth_token';

export interface AuthProfile {
	id: number;
	email: string;
	name: string;
	selectedCompanyId: number | null;
}

export interface AuthState {
	token: string;
	profile: AuthProfile;
}

export function getToken(): string | null {
	if (typeof localStorage === 'undefined') return null;
	return localStorage.getItem(TOKEN_KEY);
}

export function setAuth(state: AuthState): void {
	localStorage.setItem(TOKEN_KEY, state.token);
	localStorage.setItem('auth_profile', JSON.stringify(state.profile));
}

export function clearAuth(): void {
	localStorage.removeItem(TOKEN_KEY);
	localStorage.removeItem('auth_profile');
}

export function getProfile(): AuthProfile | null {
	if (typeof localStorage === 'undefined') return null;
	const raw = localStorage.getItem('auth_profile');
	if (!raw) return null;
	try {
		return JSON.parse(raw) as AuthProfile;
	} catch {
		return null;
	}
}

export function isAuthenticated(): boolean {
	return getToken() !== null;
}

export async function handleUnauthorized(): Promise<void> {
	clearAuth();
	await goto('/login');
}
