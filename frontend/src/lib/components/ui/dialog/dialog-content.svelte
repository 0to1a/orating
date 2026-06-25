<script lang="ts">
	import { cn } from '$lib/utils.js';
	let { class: className = '', children }: { class?: string; children?: import('svelte').Snippet } = $props();
</script>
<div class={cn('bg-card relative grid w-full max-w-lg gap-4 rounded-2xl border p-6 shadow-[0_30px_80px_rgba(0,0,0,.6)]', className)}>
	{#if children}{@render children()}{/if}
</div>
