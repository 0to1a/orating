<script lang="ts">
	import { createQuery, createMutation } from '@tanstack/svelte-query';
	import { getProfile, setAuth, getToken } from '$lib/auth.js';
	import { listCompanies, selectCompany } from '$lib/api-client.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import ChevronsUpDown from '@lucide/svelte/icons/chevrons-up-down';
	import Plus from '@lucide/svelte/icons/plus';
	import Building2 from '@lucide/svelte/icons/building-2';
	import { goto } from '$app/navigation';
	import type { CompanyInfo } from '$lib/api/types.gen.js';

	const sidebar = Sidebar.useSidebar();

	const profile = $derived(getProfile());

	const companiesQuery = createQuery({
		queryKey: ['companies'],
		queryFn: () => listCompanies().then((r) => r.data?.companies ?? [])
	});

	const activeCompany = $derived(
		$companiesQuery.data?.find((c) => c.id === profile?.selectedCompanyId) ?? null
	);

	const selectMut = createMutation({
		mutationFn: (id: number) => selectCompany({ path: { id } }),
		onSuccess: (_, id) => {
			const token = getToken()!;
			const p = getProfile()!;
			setAuth({ token, profile: { ...p, selectedCompanyId: id } });
			window.location.href = '/app/dashboard';
		}
	});
</script>

<Sidebar.Menu>
	<Sidebar.MenuItem>
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Sidebar.MenuButton
						{...props}
						size="lg"
						class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
					>
						<div class="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
							<Building2 class="size-4" />
						</div>
						<div class="grid flex-1 text-start text-sm leading-tight">
							<span class="truncate font-semibold">
								{activeCompany?.name ?? 'Select company'}
							</span>
							<div class="mt-0.5 flex items-center gap-1">
								{#if activeCompany?.role}
									<span class="role-pill capitalize">{activeCompany.role}</span>
								{/if}
							</div>
						</div>
						<ChevronsUpDown class="ms-auto" />
					</Sidebar.MenuButton>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content
				class="w-(--bits-dropdown-menu-anchor-width) min-w-56 rounded-lg"
				align="start"
				side={sidebar.isMobile ? 'bottom' : 'right'}
				sideOffset={4}
			>
				<DropdownMenu.Label class="text-muted-foreground text-xs">Companies</DropdownMenu.Label>
				{#each $companiesQuery.data ?? [] as company (company.id)}
					<DropdownMenu.Item
						onSelect={() => $selectMut.mutate(company.id)}
						class="gap-2 p-2"
					>
						<div class="flex size-6 items-center justify-center rounded-md border">
							<Building2 class="size-3.5 shrink-0" />
						</div>
						{company.name}
						{#if company.id === activeCompany?.id}
							<span class="text-muted-foreground ms-auto text-xs">active</span>
						{/if}
					</DropdownMenu.Item>
				{/each}
				<DropdownMenu.Separator />
				<DropdownMenu.Item class="gap-2 p-2" onSelect={() => goto('/app/companies')}>
					<div class="flex size-6 items-center justify-center rounded-md border bg-transparent">
						<Plus class="size-4" />
					</div>
					<div class="text-muted-foreground font-medium">Add company</div>
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.MenuItem>
</Sidebar.Menu>
