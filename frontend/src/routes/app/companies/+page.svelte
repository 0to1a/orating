<script lang="ts">
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { Plus } from 'lucide-svelte';
	import type { CompanyInfo } from '$lib/api/types.gen.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import {
		Dialog,
		DialogContent,
		DialogHeader,
		DialogTitle,
		DialogDescription,
		DialogFooter
	} from '$lib/components/ui/dialog/index.js';
	import { getProfile, getToken, setAuth } from '$lib/auth.js';
	import { listCompanies, createCompany, selectCompany } from '$lib/api-client.js';

	const qc = useQueryClient();

	const query = createQuery({
		queryKey: ['companies'],
		queryFn: () => listCompanies().then((r) => r.data?.companies ?? [])
	});

	// Local state tracks the selected company id so cards update immediately
	// after a switch without waiting for a page reload.
	let currentCompanyId = $state(getProfile()?.selectedCompanyId ?? null);

	let createOpen = $state(false);
	let companyName = $state('');

	const createMut = createMutation({
		mutationFn: () => createCompany({ body: { name: companyName } }),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: ['companies'] });
			createOpen = false;
			companyName = '';
			toast.success('Company created');
		},
		onError: (e: Error) => toast.error(e.message ?? 'Failed to create company')
	});

	const selectMut = createMutation({
		mutationFn: (id: number) => selectCompany({ path: { id } }),
		onSuccess: async (_, id) => {
			const token = getToken()!;
			const p = getProfile()!;
			setAuth({ token, profile: { ...p, selectedCompanyId: id } });
			currentCompanyId = id;
			const switched = $query.data?.find((c: CompanyInfo) => c.id === id)?.name ?? '';
			toast.success(`Switched to ${switched}`);
			// Navigate to same page so layout components re-read localStorage
			// (sidebar company name updates without a full reload)
			await goto('/app/companies');
		},
		onError: (e: Error) => toast.error(e.message ?? 'Failed to switch company')
	});

	// Deterministic avatar color by company id — 5-slot palette using existing tokens
	const AV_COLORS = [
		{ bg: 'var(--brand-soft)', color: 'var(--brand-text)' },
		{ bg: 'var(--ok-soft)', color: 'var(--ok)' },
		{ bg: 'var(--danger-soft)', color: 'var(--danger)' },
		{ bg: 'var(--warn-soft)', color: 'var(--warn)' },
		{ bg: 'var(--s3)', color: 'var(--t2)' }
	] as const;

	function avColor(id: number) {
		return AV_COLORS[Math.abs(id) % AV_COLORS.length];
	}

	function companyInitials(name: string): string {
		const parts = name.trim().split(/\s+/);
		return parts.length >= 2
			? (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
			: name.slice(0, 2).toUpperCase();
	}
</script>

<div class="px-4 py-8 md:px-6">
	<div class="mx-auto max-w-[1200px]">
		<!-- Page header -->
		<div class="mb-6 flex items-start justify-between gap-4">
			<div>
				<h1 class="text-xl font-semibold text-t1">Companies</h1>
				<p class="mt-1 text-sm text-t2">Switch between your memberships.</p>
			</div>
			<Button onclick={() => (createOpen = true)}>
				<Plus class="mr-1.5 h-4 w-4" /> New company
			</Button>
		</div>

		{#if $query.isLoading}
			<!-- Skeleton grid -->
			<div
				class="grid gap-[14px]"
				style="grid-template-columns: repeat(auto-fill, minmax(290px, 1fr))"
			>
				{#each { length: 4 } as _}
					<Skeleton class="h-[120px] rounded-xl" />
				{/each}
			</div>
		{:else if ($query.data?.length ?? 0) === 0}
			<!-- Empty state -->
			<div class="flex flex-col items-center justify-center gap-4 py-24 text-center">
				<p class="text-sm text-t2">You're not in any companies yet.</p>
				<Button onclick={() => (createOpen = true)}>
					<Plus class="mr-1.5 h-4 w-4" /> New company
				</Button>
			</div>
		{:else}
			<!-- Company card grid -->
			<div
				class="grid gap-[14px]"
				style="grid-template-columns: repeat(auto-fill, minmax(290px, 1fr))"
			>
				{#each $query.data ?? [] as company (company.id)}
					{@const av = avColor(company.id)}
					{@const isCurrent = company.id === currentCompanyId}
					<div
						class="flex flex-col gap-3 rounded-xl p-[18px]"
						style="background: var(--s2); box-shadow: 0 0 0 1px var(--border-h)"
					>
						<!-- Logo tile + name -->
						<div class="flex items-center gap-3">
							<div
								class="avatar-initials h-10 w-10 flex-shrink-0 text-sm"
								style="background: {av.bg}; color: {av.color}"
							>
								{companyInitials(company.name)}
							</div>
							<p class="min-w-0 truncate font-semibold text-t1">{company.name}</p>
						</div>

						<!-- Footer: role pill + current/switch -->
						<div class="flex items-center justify-between gap-2">
							<span class={company.isOwner ? 'role-pill' : 'role-pill muted'}>
								{company.isOwner ? 'Admin' : company.role}
							</span>

							{#if isCurrent}
								<span
									class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium"
									style="background: var(--ok-soft); color: var(--ok)"
								>
									Current
								</span>
							{:else}
								<Button
									size="sm"
									variant="outline"
									onclick={() => $selectMut.mutate(company.id)}
									disabled={$selectMut.isPending}
								>
									Switch
								</Button>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- New company dialog -->
<Dialog bind:open={createOpen}>
	<DialogContent>
		<DialogHeader>
			<DialogTitle>New company</DialogTitle>
			<DialogDescription>Create a new company workspace.</DialogDescription>
		</DialogHeader>
		<div class="space-y-2 py-2">
			<Label for="company-name">Company name</Label>
			<Input id="company-name" bind:value={companyName} placeholder="Acme Corp" />
		</div>
		<DialogFooter>
			<Button variant="outline" onclick={() => (createOpen = false)}>Cancel</Button>
			<Button
				onclick={() => $createMut.mutate()}
				disabled={$createMut.isPending || !companyName.trim()}
			>
				{$createMut.isPending ? 'Creating…' : 'Create'}
			</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>
