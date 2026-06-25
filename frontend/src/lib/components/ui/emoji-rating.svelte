<script lang="ts">
	interface EmojiItem {
		emoji: string;
		label: string;
	}

	interface Props {
		items?: EmojiItem[];
		value?: number;
		onRate?: (rating: number) => void;
		disabled?: boolean;
		class?: string;
	}

	const DEFAULT_ITEMS: EmojiItem[] = [
		{ emoji: '😔', label: 'Terrible' },
		{ emoji: '😕', label: 'Poor' },
		{ emoji: '😐', label: 'Okay' },
		{ emoji: '🙂', label: 'Good' },
		{ emoji: '😍', label: 'Amazing' },
	];

	let {
		items = DEFAULT_ITEMS,
		value = $bindable(0),
		onRate,
		disabled = false,
		class: className = '',
	}: Props = $props();

	let hover = $state(0);
	const display = $derived(hover || value);

	function handleClick(val: number) {
		if (disabled) return;
		value = val;
		onRate?.(val);
	}
</script>

<div class="emoji-rating {className}" class:disabled>
	<div class="emoji-row">
		{#each items as item, i}
			{@const val = i + 1}
			{@const active = val <= display}
			<button
				type="button"
				class="emoji-btn"
				class:active
				{disabled}
				onclick={() => handleClick(val)}
				onmouseenter={() => !disabled && (hover = val)}
				onmouseleave={() => !disabled && (hover = 0)}
				aria-label="Rate {val}: {item.label}"
				aria-pressed={value === val}
			>
				<span class="emoji-glyph" class:active>{item.emoji}</span>
			</button>
		{/each}
	</div>

	<div class="emoji-label">
		<span class="label-placeholder" class:visible={display === 0}>Rate us</span>
		{#each items as item, i}
			<span class="label-item" class:visible={display === i + 1}>{item.label}</span>
		{/each}
	</div>
</div>

<style>
	.emoji-rating {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 14px;
	}
	.emoji-rating.disabled {
		opacity: 0.5;
		pointer-events: none;
	}

	.emoji-row {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.emoji-btn {
		background: none;
		border: none;
		padding: 6px;
		cursor: pointer;
		border-radius: 12px;
		transition: transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
		transform: scale(1);
		line-height: 1;
	}
	.emoji-btn:hover:not(:disabled),
	.emoji-btn.active {
		transform: scale(1.18);
	}
	.emoji-btn:active:not(:disabled) {
		transform: scale(0.92);
	}
	.emoji-btn:focus-visible {
		outline: none;
		box-shadow: 0 0 0 3px var(--brand-soft);
	}

	.emoji-glyph {
		font-size: 34px;
		line-height: 1;
		display: block;
		filter: grayscale(1);
		opacity: 0.35;
		transition:
			filter 0.25s ease,
			opacity 0.25s ease;
		user-select: none;
	}
	.emoji-glyph.active {
		filter: grayscale(0);
		opacity: 1;
	}
	.emoji-btn:hover:not(:disabled) .emoji-glyph:not(.active) {
		opacity: 0.65;
		filter: grayscale(0.3);
	}

	.emoji-label {
		position: relative;
		height: 22px;
		width: 120px;
	}

	.label-placeholder,
	.label-item {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 13px;
		font-weight: 600;
		color: var(--t1);
		letter-spacing: 0.02em;
		white-space: nowrap;
		transition:
			opacity 0.25s ease,
			filter 0.25s ease,
			transform 0.25s ease;
		opacity: 0;
		filter: blur(6px);
		transform: scale(1.08);
	}

	.label-placeholder {
		font-weight: 500;
		color: var(--t3);
	}

	.label-placeholder.visible,
	.label-item.visible {
		opacity: 1;
		filter: blur(0);
		transform: scale(1);
	}
</style>
