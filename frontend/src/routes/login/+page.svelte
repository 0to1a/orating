<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { isAuthenticated } from '$lib/auth.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { toast } from 'svelte-sonner';

	onMount(() => {
		if (isAuthenticated()) goto('/app/dashboard');
	});

	let email = $state('');
	let loading = $state(false);

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
		<CardContent>
			<form onsubmit={handleSubmit} class="space-y-4">
				<div class="space-y-2">
					<Label for="email">Email</Label>
					<Input
						id="email"
						type="email"
						placeholder="you@example.com"
						bind:value={email}
						disabled={loading}
						data-testid="login-email-input"
					/>
				</div>
				<Button type="submit" class="w-full" disabled={loading || !email.trim()}>
					{loading ? 'Sending…' : 'Continue'}
				</Button>
			</form>
		</CardContent>
	</Card>
</div>
