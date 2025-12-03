<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let copied = false;

  const dispatch = createEventDispatcher();

  function handleClick() {
    dispatch('share');
  }
</script>

<button
  on:click={handleClick}
  class="share-button"
  aria-label="Copy shareable link"
  title="Copy shareable link (Ctrl+Shift+S)"
>
  {#if copied}
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5">
      <path fill-rule="evenodd" d="M19.916 4.626a.75.75 0 01.208 1.04l-9 13.5a.75.75 0 01-1.154.114l-6-6a.75.75 0 011.06-1.06l5.353 5.353 8.493-12.739a.75.75 0 011.04-.208z" clip-rule="evenodd" />
    </svg>
  {:else}
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5">
      <path fill-rule="evenodd" d="M19.902 4.098a3.75 3.75 0 00-5.304 0l-4.5 4.5a3.75 3.75 0 001.035 6.037.75.75 0 01-.646 1.353 5.25 5.25 0 01-1.449-8.45l4.5-4.5a5.25 5.25 0 117.424 7.424l-1.757 1.757a.75.75 0 11-1.06-1.06l1.757-1.757a3.75 3.75 0 000-5.304zm-7.389 4.267a.75.75 0 011-.353 5.25 5.25 0 011.449 8.45l-4.5 4.5a5.25 5.25 0 11-7.424-7.424l1.757-1.757a.75.75 0 111.06 1.06l-1.757 1.757a3.75 3.75 0 105.304 5.304l4.5-4.5a3.75 3.75 0 00-1.035-6.037.75.75 0 01-.354-1z" clip-rule="evenodd" />
    </svg>
  {/if}
</button>

<style>
  .share-button {
    @apply p-2 rounded-lg border-2 transition-all duration-200;
    @apply bg-transparent text-light-text-muted border-light-border;
    @apply dark:text-monokai-text-muted dark:border-monokai-border;
    @apply hover:text-light-cyan hover:border-light-cyan;
    @apply dark:hover:text-monokai-cyan dark:hover:border-monokai-cyan;
  }
</style>
