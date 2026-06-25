<script lang="ts">
	import { cn } from '$lib/utils.js';
	let { class: className = '', children }: { class?: string; children?: import('svelte').Snippet } = $props();
</script>
<div class={cn('flex flex-col space-y-1.5 text-center sm:text-left', className)}>{#if children}{@render children()}{/if}</div>
