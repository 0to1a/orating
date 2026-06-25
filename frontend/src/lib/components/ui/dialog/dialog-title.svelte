<script lang="ts">
	import { cn } from '$lib/utils.js';
	let { class: className = '', children }: { class?: string; children?: import('svelte').Snippet } = $props();
</script>
<h2 class={cn('text-lg font-semibold leading-none tracking-tight', className)}>{#if children}{@render children()}{/if}</h2>
