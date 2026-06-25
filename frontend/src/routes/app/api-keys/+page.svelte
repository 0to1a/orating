<script lang="ts">
	import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import Plus from '@lucide/svelte/icons/plus';
	import Trash2 from '@lucide/svelte/icons/trash-2';
	import Copy from '@lucide/svelte/icons/copy';
	import Key from '@lucide/svelte/icons/key';
	import X from '@lucide/svelte/icons/x';
	import TriangleAlert from '@lucide/svelte/icons/triangle-alert';
	import {
		Dialog,
		DialogContent,
		DialogHeader,
		DialogTitle,
		DialogDescription,
		DialogFooter
	} from '$lib/components/ui/dialog/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { listApiKeys, createApiKey, revokeApiKey } from '$lib/api-client.js';

	const qc = useQueryClient();

	const query = createQuery({
		queryKey: ['apikeys'],
		queryFn: () => listApiKeys().then((r) => r.data?.apiKeys ?? [])
	});

	let createOpen = $state(false);
	let keyName = $state('');
	let newKey = $state('');
	let revokeConfirmId = $state<number | null>(null);

	$effect(() => {
		if (!createOpen) {
			keyName = '';
			newKey = '';
		}
	});

	const createMut = createMutation({
		mutationFn: () => createApiKey({ body: { name: keyName } }),
		onSuccess: (r) => {
			qc.invalidateQueries({ queryKey: ['apikeys'] });
			newKey = r.data?.apiKey?.token ?? '';
		},
		onError: (e: Error) => toast.error(e.message ?? 'Failed to create key')
	});

	const revokeMut = createMutation({
		mutationFn: (id: number) => revokeApiKey({ path: { id } }),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: ['apikeys'] });
			revokeConfirmId = null;
			toast.success('API key revoked');
		},
		onError: (e: Error) => toast.error(e.message ?? 'Failed to revoke key')
	});

	function copyText(text: string, label = 'Copied') {
		navigator.clipboard.writeText(text);
		toast.success(label);
	}

	function closeCreate() {
		createOpen = false;
	}
</script>

<svelte:window
	onkeydown={(e) => {
		if (e.key === 'Escape') {
			if (createOpen) closeCreate();
			else revokeConfirmId = null;
		}
	}}
/>

<div class="flex flex-col gap-6 py-6">
	<!-- Page header -->
	<div class="flex items-end justify-between px-4 lg:px-6">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">API Keys</h1>
			<p class="mt-1 text-sm" style="color: var(--t2)">Manage programmatic access tokens.</p>
		</div>
		<button
			onclick={() => (createOpen = true)}
			class="inline-flex items-center gap-2 rounded-[9px] bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
		>
			<Plus class="size-4" />
			New key
		</button>
	</div>

	<div class="px-4 lg:px-6">
		{#if $query.isLoading}
			<Skeleton class="h-32 w-full" />
		{:else if ($query.data?.length ?? 0) === 0}
			<!-- Empty state -->
			<div
				class="rounded-[var(--radius-card)] border py-16 text-center"
				style="background: var(--s2); border-color: var(--border-h)"
			>
				<div
					class="mx-auto mb-4 grid size-12 place-items-center rounded-xl"
					style="background: var(--s3)"
				>
					<Key class="size-5" style="color: var(--t3)" />
				</div>
				<p class="text-[15px] font-semibold" style="color: var(--t1)">No API keys yet</p>
				<p class="mt-1 text-[13px]" style="color: var(--t3)">
					Create a key to authenticate programmatic requests.
				</p>
				<button
					onclick={() => (createOpen = true)}
					class="mt-5 inline-flex items-center gap-2 rounded-[9px] bg-primary px-4 py-2 text-[13px] font-medium text-primary-foreground transition-opacity hover:opacity-90"
				>
					<Plus class="size-4" />
					New key
				</button>
			</div>
		{:else}
			<!-- Keys table -->
			<div
				class="overflow-hidden rounded-[var(--radius-card)] border"
				style="background: var(--s2); border-color: var(--border-h)"
			>
				<table class="w-full border-collapse">
					<thead>
						<tr>
							{#each ['Name', 'Key', 'Created', 'Last used', ''] as col}
								<th
									class="px-[18px] py-[13px] text-left text-[11px] font-semibold uppercase tracking-[0.05em]"
									style="background: var(--s1); color: var(--t3); border-bottom: 1px solid var(--border-h)"
								>{col}</th>
							{/each}
						</tr>
					</thead>
					<tbody>
						{#each $query.data ?? [] as key (key.id)}
							<tr class="group">
								<td
									class="px-[18px] py-[14px] text-[14px] font-semibold"
									style="border-bottom: 1px solid var(--border-h); color: var(--t1)"
								>{key.name}</td>

								<td
									class="px-[18px] py-[14px]"
									style="border-bottom: 1px solid var(--border-h)"
								>
									<div class="flex items-center gap-2">
										<span class="mono text-[13px]" style="color: var(--t2)"
											>{key.prefix}••••••••</span
										>
										<button
											onclick={() => copyText(key.prefix, 'Key copied')}
											aria-label="Copy key prefix"
											class="grid size-6 place-items-center rounded opacity-0 transition-all group-hover:opacity-100 hover:bg-[var(--s4)]"
											style="color: var(--t3)"
										>
											<Copy class="size-3.5" />
										</button>
									</div>
								</td>

								<td
									class="mono px-[18px] py-[14px] text-[13px]"
									style="border-bottom: 1px solid var(--border-h); color: var(--t2)"
								>
									{new Date(key.createdAt).toLocaleDateString()}
								</td>

								<td
									class="mono px-[18px] py-[14px] text-[13px]"
									style="border-bottom: 1px solid var(--border-h); color: {key.lastUsedAt
										? 'var(--t2)'
										: 'var(--t3)'}"
								>
									{key.lastUsedAt ? new Date(key.lastUsedAt).toLocaleDateString() : 'Never'}
								</td>

								<td
									class="px-[18px] py-[14px]"
									style="border-bottom: 1px solid var(--border-h)"
								>
									{#if revokeConfirmId === key.id}
										<div class="flex items-center gap-2 text-[13px]">
											<span style="color: var(--t2)">Remove?</span>
											<button
												onclick={() => (revokeConfirmId = null)}
												class="rounded px-2.5 py-1 text-[12px] font-medium transition-colors hover:bg-[var(--s4)]"
												style="color: var(--t2)"
											>Cancel</button>
											<button
												onclick={() => $revokeMut.mutate(key.id)}
												disabled={$revokeMut.isPending}
												class="rounded px-2.5 py-1 text-[12px] font-medium transition-opacity hover:opacity-90 disabled:opacity-50"
												style="background: var(--danger-soft); color: var(--danger)"
											>{$revokeMut.isPending ? '…' : 'Remove'}</button>
										</div>
									{:else}
										<button
											onclick={() => (revokeConfirmId = key.id)}
											aria-label="Revoke API key"
											class="grid size-8 place-items-center rounded-lg border-none bg-transparent opacity-0 transition-all group-hover:opacity-100 hover:bg-[var(--danger-soft)] hover:!text-[var(--danger)]"
											style="color: var(--t3)"
										>
											<Trash2 class="size-4" />
										</button>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>

<!-- Create modal (two-step: form → key reveal) -->
<Dialog bind:open={createOpen}>
	<DialogContent class="w-[430px] max-w-full p-6">
		{#if !newKey}
			<!-- Step 1: name form -->
			<DialogHeader class="mb-5">
				<div class="flex items-start justify-between">
					<div>
						<DialogTitle class="text-[18px] font-semibold">Create API key</DialogTitle>
						<DialogDescription class="mt-1 text-[13px] text-muted-foreground">
							Give your key a descriptive name.
						</DialogDescription>
					</div>
					<button
						onclick={closeCreate}
						aria-label="Close"
						class="grid size-8 place-items-center rounded-lg border-none bg-transparent transition-colors hover:bg-[var(--s3)]"
						style="color: var(--t2)"
					>
						<X class="size-4" />
					</button>
				</div>
			</DialogHeader>

			<div class="mb-5">
				<label
					for="key-name"
					class="mb-1.5 block text-[12.5px] font-medium"
					style="color: var(--t2)"
				>Name</label>
				<input
					id="key-name"
					bind:value={keyName}
					placeholder="e.g. CI/CD pipeline"
					class="w-full rounded-[9px] border px-3 py-2.5 text-[13.5px] outline-none transition-shadow"
					style="background: var(--s3); border-color: var(--border-h2); color: var(--t1); font-family: inherit"
					onfocus={(e) => (e.currentTarget.style.boxShadow = '0 0 0 3px var(--brand-soft)')}
					onblur={(e) => (e.currentTarget.style.boxShadow = '')}
					onkeydown={(e) => {
						if (e.key === 'Enter' && keyName.trim() && !$createMut.isPending)
							$createMut.mutate();
					}}
				/>
			</div>

			<DialogFooter class="gap-2.5 sm:justify-end">
				<button
					onclick={closeCreate}
					class="rounded-[9px] border px-4 py-2 text-[13px] font-medium transition-colors hover:bg-[var(--s3)]"
					style="background: none; border-color: var(--border-h2); color: var(--t1)"
				>Cancel</button>
				<button
					onclick={() => $createMut.mutate()}
					disabled={$createMut.isPending || !keyName.trim()}
					class="rounded-[9px] bg-primary px-4 py-2 text-[13px] font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40"
				>
					{$createMut.isPending ? 'Creating…' : 'Create'}
				</button>
			</DialogFooter>
		{:else}
			<!-- Step 2: key reveal (once) -->
			<DialogHeader class="mb-4">
				<DialogTitle class="text-[18px] font-semibold">API key created</DialogTitle>
				<DialogDescription class="mt-1 text-[13px] text-muted-foreground">
					Copy this key now — it won't be shown again.
				</DialogDescription>
			</DialogHeader>

			<!-- Warning banner -->
			<div
				class="mb-4 flex items-start gap-2.5 rounded-[9px] px-3.5 py-3 text-[13px]"
				style="background: rgba(251, 191, 36, 0.1); color: var(--warn); border: 1px solid rgba(251, 191, 36, 0.2)"
			>
				<TriangleAlert class="mt-0.5 size-4 flex-shrink-0" />
				<span>Store this safely; we only show the full key once.</span>
			</div>

			<!-- Key box -->
			<div
				class="mb-5 flex items-center gap-2 rounded-[9px] px-3 py-2.5"
				style="background: var(--s1); border: 1px solid var(--border-h2)"
			>
				<code class="mono flex-1 break-all text-[12.5px]" style="color: var(--t1)">{newKey}</code>
				<button
					onclick={() => copyText(newKey, 'Key copied')}
					aria-label="Copy key"
					class="grid size-7 flex-shrink-0 place-items-center rounded transition-colors hover:bg-[var(--s3)]"
					style="color: var(--t2)"
				>
					<Copy class="size-4" />
				</button>
			</div>

			<DialogFooter class="sm:justify-end">
				<button
					onclick={closeCreate}
					class="rounded-[9px] bg-primary px-4 py-2 text-[13px] font-medium text-primary-foreground transition-opacity hover:opacity-90"
				>Done</button>
			</DialogFooter>
		{/if}
	</DialogContent>
</Dialog>
