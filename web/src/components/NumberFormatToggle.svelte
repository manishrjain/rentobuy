<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let showFullNumbers = false;

  const dispatch = createEventDispatcher();

  function toggle() {
    dispatch('toggle');
  }
</script>

<button
  on:click={toggle}
  class="number-format-toggle"
  aria-label={showFullNumbers ? 'Switch to compact numbers' : 'Switch to full numbers'}
  title={showFullNumbers ? 'Compact numbers (1K)' : 'Full numbers (1,024)'}
>
  <span class="icon-text">{showFullNumbers ? '1,024' : '1K'}</span>
</button>

<style>
  .number-format-toggle {
    @apply p-2 rounded-lg border-2 transition-all duration-200;
    @apply bg-transparent text-light-text-muted border-light-border;
    @apply dark:text-monokai-text-muted dark:border-monokai-border;
    @apply hover:text-light-pink hover:border-light-pink;
    @apply dark:hover:text-monokai-pink dark:hover:border-monokai-pink;
  }

  .icon-text {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.75rem;
    font-weight: 600;
  }
</style>
