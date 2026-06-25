<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { isAuthenticated, setAuth } from '$lib/auth.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as InputOTP from '$lib/components/ui/input-otp/index.js';
	import { toast } from 'svelte-sonner';

	onMount(() => {
		if (isAuthenticated()) goto('/app/dashboard');
		if (!email) goto('/login');
	});

	let email = $derived($page.url.searchParams.get('email') ?? '');
	let otp = $state('');
	let loading = $state(false);

	async function handleSubmit() {
		if (otp.length !== 6) return;
		loading = true;
		try {
			const res = await fetch('/api/auth/login-verify', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email, otp })
			});

			if (!res.ok) {
				const data = await res.json();
				toast.error(data.errors?.[0]?.message ?? data.title ?? 'Invalid code');
				otp = '';
				return;
			}

			const data = await res.json();
			setAuth({ token: data.token, profile: data.profile });
			await goto('/app/dashboard');
		} catch {
			toast.error('Network error');
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (otp.length === 6) handleSubmit();
	});
</script>

<svelte:head>
	<title>App | Verify OTP</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-background px-4">
	<div class="flex w-full max-w-sm flex-col items-center gap-6 text-center">
		<div class="flex flex-col gap-2">
			<h1 class="text-2xl font-bold tracking-tight">Check your email</h1>
			<p class="text-muted-foreground text-sm">
				We sent a 6-digit code to <span class="text-foreground font-medium">{email}</span>
			</p>
		</div>

		<InputOTP.Root
			maxlength={6}
			bind:value={otp}
			disabled={loading}
			data-testid="otp-input"
		>
			{#snippet children({ cells })}
				<InputOTP.Group
					class="gap-2 *:data-[slot=input-otp-slot]:h-14 *:data-[slot=input-otp-slot]:w-11 *:data-[slot=input-otp-slot]:rounded-md *:data-[slot=input-otp-slot]:border *:data-[slot=input-otp-slot]:text-xl"
				>
					{#each cells.slice(0, 3) as cell (cell)}
						<InputOTP.Slot {cell} />
					{/each}
				</InputOTP.Group>
				<InputOTP.Separator />
				<InputOTP.Group
					class="gap-2 *:data-[slot=input-otp-slot]:h-14 *:data-[slot=input-otp-slot]:w-11 *:data-[slot=input-otp-slot]:rounded-md *:data-[slot=input-otp-slot]:border *:data-[slot=input-otp-slot]:text-xl"
				>
					{#each cells.slice(3, 6) as cell (cell)}
						<InputOTP.Slot {cell} />
					{/each}
				</InputOTP.Group>
			{/snippet}
		</InputOTP.Root>

		<div class="flex w-full flex-col gap-2">
			<Button
				onclick={handleSubmit}
				disabled={loading || otp.length !== 6}
				class="w-full"
			>
				{loading ? 'Verifying…' : 'Verify'}
			</Button>
			<Button variant="ghost" class="w-full" onclick={() => goto('/login')} type="button">
				Use a different email
			</Button>
		</div>
	</div>
</div>
