<script lang="ts">
	import { cn } from '$lib/utils.js';
	let { class: className = '', children }: { class?: string; children?: import('svelte').Snippet } = $props();
</script>
<div class={cn('flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2', className)}>{#if children}{@render children()}{/if}</div>
