<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { getSavedConfigs, deleteConfig, type SavedConfig } from '../lib/storage';

  export let isOpen = false;

  const dispatch = createEventDispatcher();

  let configs: SavedConfig[] = [];
  let selectedIndex = 0;

  $: if (isOpen) {
    configs = getSavedConfigs();
    selectedIndex = 0;
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      event.preventDefault();
      event.stopPropagation();
      dispatch('close');
    } else if (event.key === 'ArrowDown') {
      event.preventDefault();
      event.stopPropagation();
      if (configs.length > 0) {
        selectedIndex = Math.min(selectedIndex + 1, configs.length - 1);
      }
    } else if (event.key === 'ArrowUp') {
      event.preventDefault();
      event.stopPropagation();
      if (selectedIndex > 0) {
        selectedIndex--;
      }
    } else if (event.key === 'Enter') {
      event.preventDefault();
      event.stopPropagation();
      handleLoad();
    } else if (event.key === 'Delete' || event.key === 'Backspace') {
      event.preventDefault();
      event.stopPropagation();
      handleDelete();
    }
  }

  function handleLoad() {
    if (configs.length === 0) return;
    const config = configs[selectedIndex];
    if (config) {
      dispatch('load', { name: config.name, data: config.data });
    }
  }

  function handleDelete() {
    if (configs.length === 0) return;
    const config = configs[selectedIndex];
    if (config && confirm(`Delete "${config.name}"?`)) {
      deleteConfig(config.name);
      configs = getSavedConfigs();
      if (selectedIndex >= configs.length) {
        selectedIndex = Math.max(0, configs.length - 1);
      }
    }
  }

  function selectConfig(index: number) {
    selectedIndex = index;
  }

  function formatDate(timestamp: number): string {
    return new Date(timestamp).toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }
</script>

<svelte:window on:keydown={isOpen ? handleKeyDown : undefined} />

{#if isOpen}
  <div class="dialog-overlay" on:click={() => dispatch('close')}>
    <div class="dialog-container" on:click|stopPropagation>
      <div class="dialog-header">
        <span class="text-light-pink dark:text-monokai-pink">$</span> load-config
      </div>

      <div class="dialog-content">
        {#if configs.length === 0}
          <div class="empty-state text-light-text-muted dark:text-monokai-text-muted">
            No saved configurations found.
          </div>
        {:else}
          <div class="config-list-header text-light-text-muted dark:text-monokai-text-muted">
            Use ↑↓ to navigate, Enter to load, Delete to remove:
          </div>
          <div class="config-list">
            {#each configs as config, i}
              <button
                type="button"
                class="config-item"
                class:selected={selectedIndex === i}
                on:click={() => selectConfig(i)}
                on:dblclick={handleLoad}
              >
                <span class="config-name">{config.name}</span>
                <span class="config-date text-light-text-muted dark:text-monokai-text-muted">
                  {formatDate(config.savedAt)}
                </span>
              </button>
            {/each}
          </div>
        {/if}
      </div>

      <div class="dialog-footer">
        <button type="button" class="dialog-btn cancel" on:click={() => dispatch('close')}>
          Cancel <span class="text-light-text-muted dark:text-monokai-text-muted">(Esc)</span>
        </button>
        {#if configs.length > 0}
          <button type="button" class="dialog-btn delete" on:click={handleDelete}>
            Delete <span class="text-light-text-muted dark:text-monokai-text-muted">(Del)</span>
          </button>
          <button type="button" class="dialog-btn load" on:click={handleLoad}>
            Load <span class="text-light-text-muted dark:text-monokai-text-muted">(Enter)</span>
          </button>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .dialog-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .dialog-container {
    @apply bg-light-bg dark:bg-black border-2 border-light-border dark:border-monokai-border;
    border-radius: 0.5rem;
    width: 100%;
    max-width: 450px;
    font-family: 'JetBrains Mono', monospace;
  }

  .dialog-header {
    @apply text-light-text dark:text-monokai-text border-b border-light-border dark:border-monokai-border;
    padding: 0.75rem 1rem;
    font-weight: 600;
    font-size: 0.875rem;
  }

  .dialog-content {
    padding: 1rem;
    max-height: 300px;
    overflow-y: auto;
  }

  .empty-state {
    text-align: center;
    padding: 2rem 1rem;
    font-size: 0.875rem;
  }

  .config-list-header {
    font-size: 0.75rem;
    margin-bottom: 0.5rem;
  }

  .config-list {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .config-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    text-align: left;
    @apply bg-transparent text-light-text dark:text-monokai-text;
    @apply border border-transparent;
    padding: 0.5rem 0.75rem;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    border-radius: 0.25rem;
  }

  .config-item:hover {
    @apply bg-light-bg-light dark:bg-monokai-bg-light;
  }

  .config-item.selected {
    @apply bg-light-bg-light dark:bg-monokai-bg-light border-light-pink dark:border-monokai-pink;
  }

  .config-name {
    flex: 1;
  }

  .config-date {
    font-size: 0.75rem;
    margin-left: 1rem;
  }

  .dialog-footer {
    @apply border-t border-light-border dark:border-monokai-border;
    padding: 0.75rem 1rem;
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }

  .dialog-btn {
    @apply text-light-text dark:text-monokai-text;
    padding: 0.5rem 1rem;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.8125rem;
    font-weight: 600;
    border-radius: 0.25rem;
    cursor: pointer;
  }

  .dialog-btn.cancel {
    @apply bg-transparent border border-light-border dark:border-monokai-border;
  }

  .dialog-btn.cancel:hover {
    @apply bg-light-bg-light dark:bg-monokai-bg-light;
  }

  .dialog-btn.delete {
    @apply bg-transparent border border-light-orange dark:border-monokai-orange text-light-orange dark:text-monokai-orange;
  }

  .dialog-btn.delete:hover {
    @apply bg-light-bg-light dark:bg-monokai-bg-light;
  }

  .dialog-btn.load {
    @apply bg-light-cyan dark:bg-monokai-cyan text-white dark:text-black border-0;
  }

  .dialog-btn.load:hover {
    opacity: 0.9;
  }
</style>
