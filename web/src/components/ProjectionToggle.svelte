<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let projectionYears = 10;

  const dispatch = createEventDispatcher();

  function handleClick() {
    dispatch('toggle');
  }
</script>

<button
  on:click={handleClick}
  class="projection-toggle"
  aria-label="Toggle projection period"
  title="Toggle between 10, 20, and 30-year projections"
>
  <span class="font-mono text-xs font-bold">{projectionYears}y</span>
</button>

<style>
  .projection-toggle {
    @apply p-2 rounded-lg border-2 transition-all duration-200;
    @apply bg-transparent text-light-text-muted border-light-border;
    @apply dark:text-monokai-text-muted dark:border-monokai-border;
    @apply hover:text-light-purple hover:border-light-purple;
    @apply dark:hover:text-monokai-purple dark:hover:border-monokai-purple;
    min-width: 2.5rem;
  }
</style>
