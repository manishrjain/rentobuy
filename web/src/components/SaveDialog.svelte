<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { getSavedConfigNames } from '../lib/storage';

  export let isOpen = false;

  const dispatch = createEventDispatcher();

  let inputRef: HTMLInputElement;
  let newName = '';
  let savedNames: string[] = [];
  let selectedIndex = -1;

  $: if (isOpen) {
    savedNames = getSavedConfigNames();
    newName = '';
    selectedIndex = -1;
  }

  onMount(() => {
    if (isOpen && inputRef) {
      inputRef.focus();
    }
  });

  $: if (isOpen && inputRef) {
    setTimeout(() => inputRef?.focus(), 10);
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      event.preventDefault();
      event.stopPropagation();
      dispatch('close');
    } else if (event.key === 'ArrowDown') {
      event.preventDefault();
      event.stopPropagation();
      if (savedNames.length > 0) {
        selectedIndex = Math.min(selectedIndex + 1, savedNames.length - 1);
        newName = savedNames[selectedIndex];
      }
    } else if (event.key === 'ArrowUp') {
      event.preventDefault();
      event.stopPropagation();
      if (selectedIndex > 0) {
        selectedIndex--;
        newName = savedNames[selectedIndex];
      } else if (selectedIndex === 0) {
        selectedIndex = -1;
        newName = '';
      }
    } else if (event.key === 'Enter') {
      event.preventDefault();
      event.stopPropagation();
      handleSave();
    }
  }

  function handleSave() {
    const name = newName.trim();
    if (!name) return;
    dispatch('save', { name });
  }

  function selectName(name: string, index: number) {
    newName = name;
    selectedIndex = index;
    inputRef?.focus();
  }
</script>

{#if isOpen}
  <div class="dialog-overlay" on:click={() => dispatch('close')} on:keydown={handleKeyDown}>
    <div class="dialog-container" on:click|stopPropagation>
      <div class="dialog-header">
        <span class="text-light-pink dark:text-monokai-pink">$</span> save-config
      </div>

      <div class="dialog-content">
        <label class="dialog-label">
          <span class="text-light-cyan dark:text-monokai-cyan">Name:</span>
          <input
            type="text"
            bind:value={newName}
            bind:this={inputRef}
            on:keydown={handleKeyDown}
            placeholder="Enter config name..."
            class="dialog-input"
          />
        </label>

        {#if savedNames.length > 0}
          <div class="saved-list">
            <div class="saved-list-header text-light-text-muted dark:text-monokai-text-muted">
              Existing configs (↑↓ to select, overwrites if same name):
            </div>
            {#each savedNames as name, i}
              <button
                type="button"
                class="saved-item"
                class:selected={selectedIndex === i}
                on:click={() => selectName(name, i)}
              >
                {name}
              </button>
            {/each}
          </div>
        {/if}
      </div>

      <div class="dialog-footer">
        <button type="button" class="dialog-btn cancel" on:click={() => dispatch('close')}>
          Cancel <span class="text-light-text-muted dark:text-monokai-text-muted">(Esc)</span>
        </button>
        <button type="button" class="dialog-btn save" on:click={handleSave} disabled={!newName.trim()}>
          Save <span class="text-light-text-muted dark:text-monokai-text-muted">(Enter)</span>
        </button>
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
    max-width: 400px;
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
  }

  .dialog-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    font-weight: 600;
  }

  .dialog-input {
    flex: 1;
    @apply bg-light-bg-light dark:bg-monokai-bg-light text-light-text dark:text-monokai-text;
    @apply border border-light-border dark:border-monokai-border;
    border-radius: 0.25rem;
    padding: 0.5rem 0.75rem;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.875rem;
    font-weight: 600;
    outline: none;
  }

  .dialog-input:focus {
    @apply border-light-pink dark:border-monokai-pink;
  }

  .dialog-input::placeholder {
    @apply text-light-text-dim dark:text-monokai-text-dim;
    font-style: italic;
  }

  .saved-list {
    margin-top: 1rem;
  }

  .saved-list-header {
    font-size: 0.75rem;
    margin-bottom: 0.5rem;
  }

  .saved-item {
    display: block;
    width: 100%;
    text-align: left;
    @apply bg-transparent text-light-text dark:text-monokai-text;
    @apply border border-transparent;
    padding: 0.375rem 0.75rem;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.8125rem;
    font-weight: 600;
    cursor: pointer;
    border-radius: 0.25rem;
  }

  .saved-item:hover {
    @apply bg-light-bg-light dark:bg-monokai-bg-light;
  }

  .saved-item.selected {
    @apply bg-light-bg-light dark:bg-monokai-bg-light border-light-pink dark:border-monokai-pink;
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

  .dialog-btn.save {
    @apply bg-light-pink dark:bg-monokai-pink text-white border-0;
  }

  .dialog-btn.save:hover {
    opacity: 0.9;
  }

  .dialog-btn.save:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
