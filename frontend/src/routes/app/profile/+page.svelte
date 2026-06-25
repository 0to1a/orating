<script lang="ts">
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { getUserMe, updateUserMe } from '$lib/api-client.js';

	const qc = useQueryClient();

	const query = createQuery({
		queryKey: ['profile'],
		queryFn: () => getUserMe().then((r) => r.data!)
	});

	let name = $state('');
	let dirty = $state(false);

	$effect(() => {
		if ($query.data && !dirty) {
			name = $query.data.name ?? '';
		}
	});

	const mutation = createMutation({
		mutationFn: () => updateUserMe({ body: { name } }),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: ['profile'] });
			dirty = false;
			toast.success('Profile saved');
		},
		onError: (e: Error) => toast.error(e.message ?? 'Failed to update profile')
	});

	function getInitials(displayName: string, email: string): string {
		const n = displayName?.trim();
		if (n) {
			const parts = n.split(/\s+/);
			return parts.length >= 2
				? (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
				: n.slice(0, 2).toUpperCase();
		}
		return email?.[0]?.toUpperCase() ?? '?';
	}
</script>

<div class="px-4 py-8 md:px-6">
	<div class="mx-auto w-full max-w-[560px]">
		<div class="mb-6">
			<h1 class="text-xl font-semibold text-t1">Profile</h1>
			<p class="mt-1 text-sm text-t2">Your account information.</p>
		</div>

		{#if $query.isLoading}
			<div
				class="overflow-hidden rounded-xl"
				style="background: var(--s2); box-shadow: 0 0 0 1px var(--border-h)"
			>
				<div class="flex items-center gap-4 px-6 py-5">
					<Skeleton class="h-12 w-12 rounded-full" />
					<div class="space-y-2">
						<Skeleton class="h-4 w-32" />
						<Skeleton class="h-3 w-48" />
					</div>
				</div>
				<div style="border-top: 1px solid var(--border-h)" class="px-6 py-10">
					<Skeleton class="h-4 w-24 mb-3" />
					<Skeleton class="h-9 w-full" />
				</div>
			</div>
		{:else if $query.data}
			{@const user = $query.data}
			<div
				class="overflow-hidden rounded-xl"
				style="background: var(--s2); box-shadow: 0 0 0 1px var(--border-h)"
			>
				<!-- Avatar + identity header -->
				<div class="flex items-center gap-4 px-6 py-5">
					<div
						class="avatar-initials h-12 w-12 text-sm"
						style="background: var(--s4); color: var(--t1)"
					>
						{getInitials(user.name, user.email)}
					</div>
					<div class="min-w-0">
						<p class="truncate font-semibold text-t1">{user.name || '—'}</p>
						<p class="mono mt-0.5 truncate text-sm text-t2">{user.email}</p>
					</div>
				</div>

				<!-- Read-only info rows -->
				<div style="border-top: 1px solid var(--border-h)">
					<div
						class="flex items-center justify-between px-6 py-4"
						style="border-bottom: 1px solid var(--border-h)"
					>
						<span class="text-sm text-t3">Email</span>
						<span class="mono text-sm text-t1">{user.email}</span>
					</div>
					<div class="flex items-center justify-between px-6 py-4">
						<span class="text-sm text-t3">Role</span>
						<span class="role-pill">Admin</span>
					</div>
				</div>

				<!-- Editable display name -->
				<div class="px-6 py-5" style="border-top: 1px solid var(--border-h)">
					<label
						for="display-name"
						class="mb-2 block text-sm font-medium text-t2"
					>
						Display name
					</label>
					<Input
						id="display-name"
						bind:value={name}
						oninput={() => (dirty = true)}
						placeholder="Your name"
					/>
					<div class="mt-4">
						<Button
							onclick={() => $mutation.mutate()}
							disabled={$mutation.isPending || !name.trim() || !dirty}
						>
							{$mutation.isPending ? 'Saving…' : 'Save changes'}
						</Button>
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>
