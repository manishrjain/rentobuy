<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let viewMode: 'cumulative' | 'yearly' = 'cumulative';
  export let disabled = false;

  const dispatch = createEventDispatcher();

  function toggle() {
    dispatch('toggle');
  }
</script>

<button
  on:click={toggle}
  class="view-mode-toggle"
  class:disabled
  {disabled}
  aria-label={viewMode === 'cumulative' ? 'Switch to yearly view' : 'Switch to cumulative view'}
  title={viewMode === 'cumulative' ? 'Cumulative totals' : 'Yearly changes'}
>
  <span class="icon-text">{viewMode === 'cumulative' ? 'Cum' : 'Yr'}</span>
</button>

<style>
  .view-mode-toggle {
    @apply p-2 rounded-lg border-2 transition-all duration-200;
    @apply bg-transparent text-light-text-muted border-light-border;
    @apply dark:text-monokai-text-muted dark:border-monokai-border;
    @apply hover:text-light-pink hover:border-light-pink;
    @apply dark:hover:text-monokai-pink dark:hover:border-monokai-pink;
  }

  .view-mode-toggle.disabled {
    @apply opacity-40 cursor-not-allowed;
    @apply hover:text-light-text-muted hover:border-light-border;
    @apply dark:hover:text-monokai-text-muted dark:hover:border-monokai-border;
  }

  .icon-text {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.75rem;
    font-weight: 600;
  }
</style>
