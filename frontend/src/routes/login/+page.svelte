<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { signInWithPopup } from 'firebase/auth';
	import { getFirebaseAuth, googleProvider } from '$lib/firebase.js';
	import { isAuthenticated, setAuth } from '$lib/auth.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import { toast } from 'svelte-sonner';

	onMount(() => {
		if (isAuthenticated()) goto('/app/dashboard');
	});

	let email = $state('');
	let loading = $state(false);
	let googleLoading = $state(false);

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!email.trim()) return;

		loading = true;
		try {
			const res = await fetch('/api/auth/login-request', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email })
			});

			if (!res.ok) {
				const data = await res.json();
				toast.error(data.errors?.[0]?.message ?? data.title ?? 'Failed to send OTP');
				return;
			}

			await goto(`/verify-otp?email=${encodeURIComponent(email)}`);
		} catch {
			toast.error('Network error');
		} finally {
			loading = false;
		}
	}

	async function handleGoogleLogin() {
		googleLoading = true;
		try {
			const fbAuth = await getFirebaseAuth();
		const result = await signInWithPopup(fbAuth, googleProvider);
			const idToken = await result.user.getIdToken();

			const res = await fetch('/api/auth/google', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ idToken })
			});

			if (!res.ok) {
				const data = await res.json();
				toast.error(data.errors?.[0]?.message ?? data.title ?? 'Google login failed');
				return;
			}

			const data = await res.json();
			setAuth({ token: data.token, profile: data.profile });
			await goto('/app/dashboard');
		} catch (err: unknown) {
			const code = (err as { code?: string }).code;
			if (code === 'auth/popup-closed-by-user' || code === 'auth/cancelled-popup-request') return;
			toast.error('Google login failed');
		} finally {
			googleLoading = false;
		}
	}
</script>

<svelte:head>
	<title>App | Login</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center">
	<Card class="w-full max-w-sm">
		<CardHeader>
			<CardTitle>Sign in</CardTitle>
			<CardDescription>Enter your email to receive a one-time code</CardDescription>
		</CardHeader>
		<CardContent class="space-y-4">
			<Button
				variant="outline"
				class="w-full"
				onclick={handleGoogleLogin}
				disabled={googleLoading || loading}
			>
				{#if googleLoading}
					Signing in…
				{:else}
					<svg class="mr-2 h-4 w-4" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
						<path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
						<path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
						<path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" fill="#FBBC05"/>
						<path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
					</svg>
					Continue with Google
				{/if}
			</Button>

			<div class="relative">
				<div class="absolute inset-0 flex items-center">
					<Separator />
				</div>
				<div class="relative flex justify-center text-xs uppercase">
					<span class="bg-card px-2 text-muted-foreground">or</span>
				</div>
			</div>

			<form onsubmit={handleSubmit} class="space-y-4">
				<div class="space-y-2">
					<Label for="email">Email</Label>
					<Input
						id="email"
						type="email"
						placeholder="you@example.com"
						bind:value={email}
						disabled={loading || googleLoading}
						data-testid="login-email-input"
					/>
				</div>
				<Button type="submit" class="w-full" disabled={loading || googleLoading || !email.trim()}>
					{loading ? 'Sending…' : 'Continue with email'}
				</Button>
			</form>
		</CardContent>
	</Card>
</div>
