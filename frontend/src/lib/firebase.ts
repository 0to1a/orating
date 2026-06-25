import { initializeApp } from 'firebase/app';
import { getAuth, GoogleAuthProvider, type Auth } from 'firebase/auth';

export const googleProvider = new GoogleAuthProvider();

let authInstance: Auth | null = null;

export async function getFirebaseAuth(): Promise<Auth> {
	if (authInstance) return authInstance;

	const res = await fetch('/api/config');
	if (!res.ok) throw new Error('Failed to load app config');
	const cfg = await res.json();

	const app = initializeApp({
		apiKey: cfg.firebaseApiKey,
		authDomain: cfg.firebaseAuthDomain,
		projectId: cfg.firebaseProjectId
	});
	authInstance = getAuth(app);
	return authInstance;
}
