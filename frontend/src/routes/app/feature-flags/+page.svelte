<script lang="ts">
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import type { AdminFlag } from '$lib/api/types.gen.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';
	import { getProfile } from '$lib/auth.js';
	import { listFeatureFlags, adminListFeatureFlags, adminUpsertFeatureFlags } from '$lib/api-client.js';

	const qc = useQueryClient();
	const profile = getProfile();
	const isAdminCompany = $derived((profile?.selectedCompanyId ?? 0) === 1);

	const flagsQuery = createQuery({
		queryKey: ['feature-flags', 'user'],
		queryFn: () => listFeatureFlags().then((r) => r.data?.flags ?? {}),
		get enabled() { return !isAdminCompany; }
	});

	const adminQuery = createQuery({
		queryKey: ['feature-flags', 'admin'],
		queryFn: () => adminListFeatureFlags().then((r) => r.data?.flags ?? []),
		get enabled() { return isAdminCompany; }
	});

	const toggleMut = createMutation({
		mutationFn: ({ flagKey, enabled, companyId }: { flagKey: string; enabled: boolean; companyId: number }) =>
			adminUpsertFeatureFlags({ body: { flagKey, enabled, companyIds: [companyId], allCompanies: false } }),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: ['feature-flags'] });
			toast.success('Flag updated');
		},
		onError: (e: Error) => toast.error(e.message ?? 'Failed to update flag')
	});

	const isLoading = $derived(isAdminCompany ? $adminQuery.isLoading : $flagsQuery.isLoading);

	const flagEntries = $derived(
		isAdminCompany
			? ($adminQuery.data ?? [])
			: Object.entries($flagsQuery.data ?? {}).map(([key, enabled]) => ({
					flagKey: key,
					enabled,
					companyId: 0,
					companyName: '',
					id: 0
				}))
	);
</script>

<div class="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
	<div class="px-4 lg:px-6">
		<p class="text-muted-foreground text-sm">
			{isAdminCompany
				? 'Admin: manage flags across all companies'
				: 'Feature flags for your company'}
		</p>
	</div>

	<div class="px-4 lg:px-6">
		{#if isLoading}
			<Skeleton class="h-32 w-full" />
		{:else if flagEntries.length === 0}
			<p class="text-muted-foreground text-sm">No feature flags configured.</p>
		{:else}
			<Table>
				<TableHeader>
					<TableRow>
						<TableHead>Flag key</TableHead>
						{#if isAdminCompany}<TableHead>Company</TableHead>{/if}
						<TableHead>Status</TableHead>
						{#if isAdminCompany}<TableHead>Action</TableHead>{/if}
					</TableRow>
				</TableHeader>
				<TableBody>
					{#each flagEntries as flag}
						<TableRow>
							<TableCell class="font-mono text-sm">{flag.flagKey}</TableCell>
							{#if isAdminCompany}
								<TableCell class="text-muted-foreground text-sm">
									{(flag as AdminFlag).companyName}
								</TableCell>
							{/if}
							<TableCell>
								<Badge variant={flag.enabled ? 'default' : 'secondary'}>
									{flag.enabled ? 'Enabled' : 'Disabled'}
								</Badge>
							</TableCell>
							{#if isAdminCompany}
								<TableCell>
									<Button
										variant="outline"
										size="sm"
										onclick={() =>
											$toggleMut.mutate({
												flagKey: flag.flagKey,
												enabled: !flag.enabled,
												companyId: (flag as AdminFlag).companyId
											})}
										disabled={$toggleMut.isPending}
									>
										{flag.enabled ? 'Disable' : 'Enable'}
									</Button>
								</TableCell>
							{/if}
						</TableRow>
					{/each}
				</TableBody>
			</Table>
		{/if}
	</div>
</div>
